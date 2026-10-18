package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

func TestScopedTransactionsIsolationRollbackCopiesAndIndex(t *testing.T) {
	ctx := context.Background()
	m := NewMemoryEvents(domain.DefaultRuleConfig())
	if err := m.Update(ctx, func(s *EventState) error {
		s.Events["a"] = EventRecord{Input: EventInput{Account: "a"}, Result: domain.RiskDecision{ReasonCodes: []string{"original"}}}
		s.Events["b"] = EventRecord{Input: EventInput{Account: "b"}}
		s.Feedback["b"] = "FALSE_POSITIVE"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	scope := EventScope{EventID: "new", Account: "a", History: true}
	sentinel := errors.New("rollback")
	if err := m.UpdateScoped(ctx, scope, func(s *EventState) error {
		if len(s.Events) != 1 || len(s.Feedback) != 0 {
			t.Fatal("unrelated history/feedback copied")
		}
		delete(s.Events, "a")
		s.Rules.ReviewAt = 1
		return sentinel
	}); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	var retained *EventState
	if err := m.UpdateScoped(ctx, scope, func(s *EventState) error {
		if s.Events["a"].Input.Account != "a" || s.Rules.ReviewAt != 40 {
			t.Fatal("rollback failed")
		}
		s.Events["new"] = EventRecord{Input: EventInput{Account: "a"}, Result: domain.RiskDecision{ReasonCodes: []string{"new"}}}
		s.Feedback["new"] = "CONFIRMED_FRAUD"
		retained = s
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	retained.Events["a"].Result.ReasonCodes[0] = "corrupted"
	retained.Events["new"].Result.ReasonCodes[0] = "corrupted"
	retained.Feedback["new"] = "bad"
	if err := m.UpdateScoped(ctx, EventScope{Account: "a", History: true}, func(s *EventState) error {
		if len(s.Events) != 2 || s.Events["a"].Result.ReasonCodes[0] != "original" || s.Events["new"].Result.ReasonCodes[0] != "new" {
			t.Fatal("index/copies broken")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := m.Update(ctx, func(s *EventState) error {
		if len(s.Events) != 3 || s.Feedback["b"] != "FALSE_POSITIVE" || s.Feedback["new"] != "CONFIRMED_FRAUD" {
			t.Fatal("unrelated state lost")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// Moving the addressed event updates both account indexes; deletion cleans up.
	if err := m.UpdateScoped(ctx, EventScope{EventID: "new"}, func(s *EventState) error {
		e := s.Events["new"]
		e.Input.Account = "c"
		s.Events["new"] = e
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := m.UpdateScoped(ctx, EventScope{Account: "c", History: true}, func(s *EventState) error {
		if len(s.Events) != 1 {
			t.Fatal("move index missing")
		}
		delete(s.Events, "new")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(m.byAccount["c"]) != 0 {
		t.Fatal("empty index retained")
	}
}

func TestScopedTransactionsRejectEscapesAndCancelAtomically(t *testing.T) {
	ctx := context.Background()
	m := NewMemoryEvents(domain.DefaultRuleConfig())
	_ = m.Update(ctx, func(s *EventState) error {
		s.Events["hidden"] = EventRecord{Input: EventInput{Account: "other"}}
		return nil
	})
	scope := EventScope{EventID: "new", Account: "a", History: true}
	for _, fn := range []func(*EventState) error{
		func(s *EventState) error {
			s.Events["outside"] = EventRecord{Input: EventInput{Account: "other"}}
			return nil
		},
		func(s *EventState) error {
			s.Events["hidden"] = EventRecord{Input: EventInput{Account: "a"}}
			return nil
		},
		func(s *EventState) error { s.Feedback["outside"] = "bad"; return nil },
		func(s *EventState) error { s.Rules.ReviewAt = 0; return nil },
	} {
		if err := m.UpdateScoped(ctx, scope, fn); err == nil {
			t.Fatal("invalid scoped write committed")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := m.UpdateScoped(canceled, scope, func(*EventState) error { t.Fatal("canceled callback invoked"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	mid, cancelMid := context.WithCancel(ctx)
	if err := m.UpdateScoped(mid, scope, func(s *EventState) error { s.Rules.ReviewAt = 1; cancelMid(); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := m.UpdateScoped(ctx, EventScope{}, func(s *EventState) error {
		if len(s.Events) != 0 || len(s.Feedback) != 0 || s.Rules.ReviewAt != 40 {
			t.Fatal("policy view/rollback broken")
		}
		s.Feedback["new"] = "bad"
		return nil
	}); err == nil {
		t.Fatal("policy-only scope accepted feedback")
	}
	_ = m.Update(ctx, func(s *EventState) error {
		if len(s.Events) != 1 || s.Events["hidden"].Input.Account != "other" || len(s.Feedback) != 0 {
			t.Fatal("failed scope mutated state")
		}
		return nil
	})
}
