package query

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
)

func TestVelocityWindowReplayAndAccountIsolation(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	config := domain.DefaultRuleConfig()
	config.VelocityLimit = 2
	store := repository.NewMemoryEvents(config)
	h, err := NewScoreEventHandlerWithStore(store, func() time.Time { return now }, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	score := func(id, account string) domain.RiskDecision {
		t.Helper()
		d, err := h.Handle(ctx, ScoreEventQuery{EventID: id, Account: account, AmountCents: 1})
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	first := score("first", "a")
	for i := 0; i < 10; i++ {
		if d := score("first", "a"); !reflect.DeepEqual(d, first) {
			t.Fatal("replay changed")
		}
	}
	if d := score("second", "a"); d.Score != 0 {
		t.Fatal("replay incremented velocity")
	}
	if d := score("third", "a"); d.Score != config.VelocityWeight {
		t.Fatalf("missing velocity: %+v", d)
	}
	if d := score("isolated", "b"); d.Score != 0 {
		t.Fatal("accounts leaked")
	}
	now = now.Add(time.Minute)
	if d := score("boundary", "a"); d.Score != 0 {
		t.Fatal("exclusive lower window boundary violated")
	}
	// Replay remains the original even after the window and feedback change.
	if err := h.Feedback("first", "CONFIRMED_FRAUD"); err != nil {
		t.Fatal(err)
	}
	if d := score("first", "a"); !reflect.DeepEqual(d, first) {
		t.Fatal("original explanation changed")
	}
	// Future-dated events after a clock rollback do not count.
	now = now.Add(-2 * time.Minute)
	if d := score("rollback", "a"); d.Score != 0 {
		t.Fatal("future event counted")
	}
}

func TestSharedRepositoryConcurrentUniqueEventsAndReplay(t *testing.T) {
	ctx := context.Background()
	config := domain.DefaultRuleConfig()
	config.VelocityLimit = 1
	store := repository.NewMemoryEvents(config)
	now := func() time.Time { return time.Unix(100, 0) }
	h1, _ := NewScoreEventHandlerWithStore(store, now, time.Minute)
	h2, _ := NewScoreEventHandlerWithStore(store, now, time.Minute)
	var wg sync.WaitGroup
	results := make(chan int, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := h1
			if i%2 == 0 {
				h = h2
			}
			d, err := h.Handle(ctx, ScoreEventQuery{EventID: fmt.Sprint(i), Account: "a"})
			if err != nil {
				t.Error(err)
				return
			}
			results <- d.Score
		}(i)
	}
	wg.Wait()
	close(results)
	zero := 0
	for score := range results {
		if score == 0 {
			zero++
		} else if score != config.VelocityWeight {
			t.Errorf("unexpected score %d", score)
		}
	}
	if zero != 1 {
		t.Fatalf("want exactly one initial event, got %d", zero)
	}
	// Another handler sharing the repository sees original decisions and feedback.
	if _, ok := h2.Explain("0"); !ok {
		t.Fatal("shared explanation absent")
	}
	if err := h1.Feedback("0", "FALSE_POSITIVE"); err != nil {
		t.Fatal(err)
	}
	if h2.Rules().ReviewAt != 41 {
		t.Fatal("feedback not shared")
	}
}

func TestDecisionCopiesAndInvalidInputs(t *testing.T) {
	h := NewScoreEventHandler()
	q := ScoreEventQuery{EventID: "e", Account: "a", AmountCents: 600000}
	d, err := h.Handle(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	d.ReasonCodes[0] = "corrupted"
	explained, ok := h.Explain("e")
	if !ok || explained.ReasonCodes[0] != "HIGH_AMOUNT" {
		t.Fatal("returned slice mutated store")
	}
	explained.ReasonCodes[0] = "corrupted"
	again, _ := h.Handle(context.Background(), q)
	if again.ReasonCodes[0] != "HIGH_AMOUNT" {
		t.Fatal("explanation slice mutated store")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.Handle(ctx, q); err == nil {
		t.Fatal("canceled context accepted")
	}
	for _, q := range []ScoreEventQuery{{}, {EventID: "x", Account: "a", AmountCents: -1}, {EventID: "x", Account: "a", RecentCount: -1}} {
		if _, err := h.Handle(context.Background(), q); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
	for _, window := range []time.Duration{0, -1} {
		if _, err := NewScoreEventHandlerWithStore(repository.NewMemoryEvents(domain.DefaultRuleConfig()), time.Now, window); err == nil {
			t.Fatal("invalid window accepted")
		}
	}
	if _, err := NewScoreEventHandlerWithStore(nil, time.Now, time.Minute); err == nil {
		t.Fatal("nil store")
	}
	if _, err := NewScoreEventHandlerWithStore(repository.NewMemoryEvents(domain.DefaultRuleConfig()), nil, time.Minute); err == nil {
		t.Fatal("nil clock")
	}
}

func TestFeedbackNeverInvalidatesRuleThresholds(t *testing.T) {
	config := domain.DefaultRuleConfig()
	config.ReviewAt = config.DenyAt
	h := NewScoreEventHandlerWithConfig(config)
	_, _ = h.Handle(context.Background(), ScoreEventQuery{EventID: "e", Account: "a"})
	if err := h.Feedback("e", "FALSE_POSITIVE"); err != nil {
		t.Fatal(err)
	}
	if err := h.Rules().Validate(); err != nil {
		t.Fatal(err)
	}
	config.ReviewAt = 1
	h = NewScoreEventHandlerWithConfig(config)
	_, _ = h.Handle(context.Background(), ScoreEventQuery{EventID: "e", Account: "a"})
	if err := h.Feedback("e", "CONFIRMED_FRAUD"); err != nil {
		t.Fatal(err)
	}
	if h.Rules().ReviewAt != 1 {
		t.Fatal("lower threshold bound violated")
	}
}
