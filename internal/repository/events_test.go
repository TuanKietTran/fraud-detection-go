package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

func TestMemoryEventsAtomicRollbackCancellationAndCopies(t *testing.T) {
	store := NewMemoryEvents(domain.DefaultRuleConfig())
	ctx := context.Background()
	sentinel := errors.New("rollback")
	if err := store.Update(ctx, func(s *EventState) error { s.Feedback["x"] = "bad"; s.Rules.ReviewAt = 1; return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := store.Update(canceled, func(*EventState) error { t.Fatal("called canceled transaction"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	err := store.Update(ctx, func(s *EventState) error {
		if len(s.Feedback) != 0 || s.Rules.ReviewAt != 40 {
			t.Fatal("rollback failed")
		}
		s.Events["e"] = EventRecord{Result: domain.RiskDecision{ReasonCodes: []string{"original"}}}
		s.Feedback["e"] = "FALSE_POSITIVE"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var retained *EventState
	if err := store.Update(ctx, func(s *EventState) error { retained = s; return nil }); err != nil {
		t.Fatal(err)
	}
	retained.Events["e"].Result.ReasonCodes[0] = "corrupted"
	retained.Feedback["e"] = "bad"
	if err := store.Update(ctx, func(s *EventState) error {
		if s.Events["e"].Result.ReasonCodes[0] != "original" || s.Feedback["e"] != "FALSE_POSITIVE" {
			t.Fatal("retained state mutated repository")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	mid, cancelMid := context.WithCancel(ctx)
	if err := store.Update(mid, func(s *EventState) error { s.Rules.ReviewAt = 1; cancelMid(); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := store.Update(ctx, func(s *EventState) error {
		if s.Rules.ReviewAt != 40 {
			t.Fatal("canceled transaction committed")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
