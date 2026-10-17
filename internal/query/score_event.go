package query

import (
	"context"
	"errors"
	"sync"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

type ScoreEventQuery struct {
	EventID                                            string
	AmountCents                                        int64
	Account                                            string
	RecentCount                                        int
	GeoMismatch, DeviceMismatch, NewPayee, Structuring bool
}
type replayRecord struct {
	query  ScoreEventQuery
	result domain.RiskDecision
}
type ScoreEventHandler struct {
	mu       sync.Mutex
	replay   map[string]replayRecord
	config   domain.RuleConfig
	feedback map[string]string
}

func NewScoreEventHandler() *ScoreEventHandler {
	return NewScoreEventHandlerWithConfig(domain.DefaultRuleConfig())
}
func NewScoreEventHandlerWithConfig(c domain.RuleConfig) *ScoreEventHandler {
	return &ScoreEventHandler{replay: make(map[string]replayRecord), config: c, feedback: make(map[string]string)}
}
func (h *ScoreEventHandler) Rules() domain.RuleConfig {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.config
}
func (h *ScoreEventHandler) Explain(eventID string) (domain.RiskDecision, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.replay[eventID]
	return r.result, ok
}
func (h *ScoreEventHandler) Feedback(eventID, label string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.replay[eventID]; !ok {
		return errors.New("event not found")
	}
	if label != "CONFIRMED_FRAUD" && label != "FALSE_POSITIVE" {
		return errors.New("invalid feedback label")
	}
	if old, ok := h.feedback[eventID]; ok {
		if old == label {
			return nil
		}
		return errors.New("feedback conflict")
	}
	h.feedback[eventID] = label
	if label == "CONFIRMED_FRAUD" && h.config.ReviewAt > 1 {
		h.config.ReviewAt--
	} else if label == "FALSE_POSITIVE" && h.config.ReviewAt < 100 {
		h.config.ReviewAt++
	}
	return nil
}
func (h *ScoreEventHandler) Handle(ctx context.Context, q ScoreEventQuery) (domain.RiskDecision, error) {
	if err := ctx.Err(); err != nil {
		return domain.RiskDecision{}, err
	}
	if q.EventID == "" || q.Account == "" || q.AmountCents < 0 {
		return domain.RiskDecision{}, errors.New("event_id, account and nonnegative amount required")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.replay[q.EventID]; ok {
		if old.query != q {
			return domain.RiskDecision{}, errors.New("event_id replay payload conflict")
		}
		return old.result, nil
	}
	score, decision, reasons, err := domain.ScoreWithRules(domain.Signals{AmountCents: q.AmountCents, RecentCount: q.RecentCount, GeoMismatch: q.GeoMismatch, DeviceMismatch: q.DeviceMismatch, NewPayee: q.NewPayee, Structuring: q.Structuring}, h.config)
	if err != nil {
		return domain.RiskDecision{}, err
	}
	result := domain.RiskDecision{EventID: q.EventID, Score: score, Decision: decision, ReasonCodes: reasons}
	h.replay[q.EventID] = replayRecord{query: q, result: result}
	return result, nil
}
