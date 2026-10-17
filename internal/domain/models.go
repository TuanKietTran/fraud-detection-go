package domain

import "github.com/google/uuid"

type Decision string

const (
	Allow  Decision = "ALLOW"
	Review Decision = "REVIEW"
	Block  Decision = "BLOCK"
)

type RiskDecision struct {
	ID         uuid.UUID
	EventID    string
	Score      int
	Decision   Decision
	ReasonCodes []string
}
