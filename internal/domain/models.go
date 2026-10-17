package domain

import "github.com/google/uuid"

type Decision string

const (
	Allow  Decision = "ALLOW"
	Review Decision = "REVIEW"
	Deny   Decision = "DENY"
	Block  Decision = "BLOCK" // Legacy alias retained for existing clients.
)

type RiskDecision struct {
	ID          uuid.UUID
	EventID     string
	Score       int
	Decision    Decision
	ReasonCodes []string
}

// ScoreAmountCents applies deterministic USD-cent thresholds. Limits are strict.
func ScoreAmountCents(amount int64) (int, Decision, []string) {
	score := 0
	reasons := make([]string, 0, 2)
	if amount > 500000 {
		score += 40
		reasons = append(reasons, "HIGH_AMOUNT")
	}
	if amount > 1000000 {
		score += 30
		reasons = append(reasons, "VERY_HIGH_AMOUNT")
	}
	decision := Allow
	if score >= 70 {
		decision = Deny
	} else if score >= 40 {
		decision = Review
	}
	return score, decision, reasons
}
