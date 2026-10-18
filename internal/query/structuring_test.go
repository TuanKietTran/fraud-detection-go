package query

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
)

func TestInferredStructuringReplayWindowAndIsolation(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprint("full_snapshot=", legacy), func(t *testing.T) {
			config := domain.DefaultRuleConfig()
			config.VelocityLimit = 100
			memory := repository.NewMemoryEvents(config)
			var store repository.EventStore = memory
			if legacy {
				store = fullSnapshotOnly{memory}
			}
			now := time.Unix(1000, 0)
			h, _ := NewScoreEventHandlerWithStore(store, func() time.Time { return now }, time.Minute)
			score := func(id, account string, amount int64) domain.RiskDecision {
				t.Helper()
				d, err := h.Handle(context.Background(), ScoreEventQuery{EventID: id, Account: account, AmountCents: amount})
				if err != nil {
					t.Fatal(err)
				}
				return d
			}
			first := score("one", "a", 90000)
			for i := 0; i < 8; i++ {
				if d := score("one", "a", 90000); !reflect.DeepEqual(d, first) {
					t.Fatal("replay changed")
				}
			}
			if d := score("two", "a", 100000); d.Score != 0 {
				t.Fatalf("premature structuring %+v", d)
			}
			third := score("three", "a", 99999)
			if third.Score != config.StructuringWeight || !reflect.DeepEqual(third.ReasonCodes, []string{"STRUCTURING"}) {
				t.Fatalf("missing pattern %+v", third)
			}
			if d := score("other", "b", 99999); d.Score != 0 {
				t.Fatal("cross-account pattern")
			}
			if d := score("low", "a", 89999); d.Score != 0 {
				t.Fatal("below-band pattern")
			}
			if d := score("high", "a", 100001); d.Score != 0 {
				t.Fatal("above-band pattern")
			}
			now = now.Add(time.Minute)
			if d := score("expired", "a", 99999); d.Score != 0 {
				t.Fatal("inclusive lower window counted")
			}
			if d := score("three", "a", 99999); !reflect.DeepEqual(d, third) {
				t.Fatal("replay lost original structuring")
			}
			now = now.Add(-2 * time.Minute)
			if d := score("rollback", "a", 99999); d.Score != 0 {
				t.Fatal("future timestamps counted")
			}
			if err := memory.Update(context.Background(), func(s *repository.EventState) error {
				if s.Events["three"].StructuringCount != 2 {
					t.Fatal("missing snapshot pattern count")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type fullSnapshotOnly struct{ store repository.EventStore }

func (s fullSnapshotOnly) Update(ctx context.Context, fn func(*repository.EventState) error) error {
	return s.store.Update(ctx, fn)
}
