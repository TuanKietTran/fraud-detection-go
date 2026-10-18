package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
)

var (
	ErrInvalid  = errors.New("invalid request")
	ErrConflict = errors.New("payload conflict")
	ErrNotFound = errors.New("event not found")
)

type ScoreEventQuery = repository.EventInput

type ScoreEventHandler struct {
	store  repository.EventStore
	now    func() time.Time
	window time.Duration
}

func NewScoreEventHandler() *ScoreEventHandler {
	return NewScoreEventHandlerWithConfig(domain.DefaultRuleConfig())
}
func NewScoreEventHandlerWithConfig(c domain.RuleConfig) *ScoreEventHandler {
	h, err := NewScoreEventHandlerWithStore(repository.NewMemoryEvents(c), time.Now, time.Minute)
	if err != nil {
		panic(err)
	}
	return h
}

// NewScoreEventHandlerWithStore supports shared repositories and deterministic clocks.
// Velocity counts prior unique events in (now-window, now], for the same account.
func NewScoreEventHandlerWithStore(store repository.EventStore, now func() time.Time, window time.Duration) (*ScoreEventHandler, error) {
	if store == nil || now == nil || window <= 0 {
		return nil, errors.New("store, clock and positive velocity window required")
	}
	return &ScoreEventHandler{store: store, now: now, window: window}, nil
}

func cloneDecision(d domain.RiskDecision) domain.RiskDecision {
	d.ReasonCodes = append([]string{}, d.ReasonCodes...)
	return d
}

func (h *ScoreEventHandler) Rules() domain.RuleConfig {
	rules, _ := h.LoadRules(context.Background())
	return rules
}
func (h *ScoreEventHandler) LoadRules(ctx context.Context) (domain.RuleConfig, error) {
	var rules domain.RuleConfig
	err := h.store.Update(ctx, func(s *repository.EventState) error { rules = s.Rules; return nil })
	return rules, err
}
func (h *ScoreEventHandler) Explain(eventID string) (domain.RiskDecision, bool) {
	result, ok, _ := h.ExplainContext(context.Background(), eventID)
	return result, ok
}
func (h *ScoreEventHandler) ExplainContext(ctx context.Context, eventID string) (domain.RiskDecision, bool, error) {
	var result domain.RiskDecision
	var ok bool
	err := h.store.Update(ctx, func(s *repository.EventState) error {
		record, found := s.Events[eventID]
		ok = found
		result = cloneDecision(record.Result)
		return nil
	})
	return result, ok, err
}
func (h *ScoreEventHandler) Feedback(eventID, label string) error {
	return h.FeedbackContext(context.Background(), eventID, label)
}
func (h *ScoreEventHandler) FeedbackContext(ctx context.Context, eventID, label string) error {
	return h.store.Update(ctx, func(s *repository.EventState) error {
		if _, ok := s.Events[eventID]; !ok {
			return ErrNotFound
		}
		if label != "CONFIRMED_FRAUD" && label != "FALSE_POSITIVE" {
			return fmt.Errorf("%w: invalid feedback label", ErrInvalid)
		}
		if old, ok := s.Feedback[eventID]; ok {
			if old == label {
				return nil
			}
			return fmt.Errorf("%w: feedback conflict", ErrConflict)
		}
		s.Feedback[eventID] = label
		if label == "CONFIRMED_FRAUD" && s.Rules.ReviewAt > 1 {
			s.Rules.ReviewAt--
		} else if label == "FALSE_POSITIVE" && s.Rules.ReviewAt < s.Rules.DenyAt {
			s.Rules.ReviewAt++
		}
		return nil
	})
}
func (h *ScoreEventHandler) Handle(ctx context.Context, q ScoreEventQuery) (domain.RiskDecision, error) {
	if err := ctx.Err(); err != nil {
		return domain.RiskDecision{}, err
	}
	if q.EventID == "" || q.Account == "" || q.AmountCents < 0 || q.RecentCount < 0 {
		return domain.RiskDecision{}, fmt.Errorf("%w: event_id, account and nonnegative amount/count required", ErrInvalid)
	}
	var result domain.RiskDecision
	err := h.store.Update(ctx, func(s *repository.EventState) error {
		if old, ok := s.Events[q.EventID]; ok {
			if old.Input != q {
				return fmt.Errorf("%w: event_id replay payload conflict", ErrConflict)
			}
			result = cloneDecision(old.Result)
			return nil
		}
		now := h.now().UTC()
		recent := 0
		for _, record := range s.Events {
			if record.Input.Account == q.Account && record.RecordedAt.After(now.Add(-h.window)) && !record.RecordedAt.After(now) {
				recent++
			}
		}
		// Preserve the published legacy signal, but clients cannot lower our count.
		if q.RecentCount > recent {
			recent = q.RecentCount
		}
		score, decision, reasons, err := domain.ScoreWithRules(domain.Signals{AmountCents: q.AmountCents, RecentCount: recent, GeoMismatch: q.GeoMismatch, DeviceMismatch: q.DeviceMismatch, NewPayee: q.NewPayee, Structuring: q.Structuring}, s.Rules)
		if err != nil {
			return err
		}
		result = domain.RiskDecision{EventID: q.EventID, Score: score, Decision: decision, ReasonCodes: reasons}
		s.Events[q.EventID] = repository.EventRecord{Input: q, Result: cloneDecision(result), RecordedAt: now, RecentCount: recent, Rules: s.Rules}
		return nil
	})
	return result, err
}
