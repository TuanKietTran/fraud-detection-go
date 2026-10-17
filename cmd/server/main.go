package main

import (
	"context"
	"fmt"

	"github.com/example/fraud-detection/internal/query"
)

func main() {
	h := query.NewScoreEventHandler()
	res, _ := h.Handle(context.Background(), query.ScoreEventQuery{
		EventID: "evt-1",
		Amount:  7500,
		Account: "acc-123",
	})
	fmt.Printf("Decision: %s Score: %d Reasons: %v\n", res.Decision, res.Score, res.ReasonCodes)
}
