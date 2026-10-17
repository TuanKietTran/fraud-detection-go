package query

import (
	"context"

	"github.com/example/fraud-detection/internal/domain"
)

type ScoreEventQuery struct {
	EventID string
	Amount  float64
	Account string
}

type ScoreEventHandler struct{}

func NewScoreEventHandler() *ScoreEventHandler { return &ScoreEventHandler{} }

func (h *ScoreEventHandler) Handle(ctx context.Context, q ScoreEventQuery) (domain.RiskDecision, error) {
	score := 0
	reasons := []string{}

	if q.Amount > 5000 {
		score += 40
		reasons = append(reasons, "HIGH_AMOUNT")
	}
	if q.Amount > 10000 {
		score += 30
		reasons = append(reasons, "VERY_HIGH_AMOUNT")
	}

	decision := domain.Allow
	if score >= 70 {
		decision = domain.Block
	} else if score >= 40 {
		decision = domain.Review
	}

	return domain.RiskDecision{
		EventID:     q.EventID,
		Score:       score,
		Decision:    decision,
		ReasonCodes: reasons,
	}, nil
}
