package query

import (
	"context"
	"errors"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

type ScoreEventQuery struct {
	EventID     string
	AmountCents int64
	Account     string
}
type ScoreEventHandler struct{}

func NewScoreEventHandler() *ScoreEventHandler { return &ScoreEventHandler{} }
func (h *ScoreEventHandler) Handle(ctx context.Context, q ScoreEventQuery) (domain.RiskDecision, error) {
	if err := ctx.Err(); err != nil {
		return domain.RiskDecision{}, err
	}
	if q.EventID == "" || q.Account == "" || q.AmountCents < 0 {
		return domain.RiskDecision{}, errors.New("event_id, account and nonnegative amount required")
	}
	score, decision, reasons := domain.ScoreAmountCents(q.AmountCents)
	return domain.RiskDecision{EventID: q.EventID, Score: score, Decision: decision, ReasonCodes: reasons}, nil
}
