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
	mu     sync.Mutex
	replay map[string]replayRecord
	config domain.RuleConfig
}

func NewScoreEventHandler() *ScoreEventHandler {
	return NewScoreEventHandlerWithConfig(domain.DefaultRuleConfig())
}
func NewScoreEventHandlerWithConfig(c domain.RuleConfig) *ScoreEventHandler {
	return &ScoreEventHandler{replay: make(map[string]replayRecord), config: c}
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
