package domain

import "testing"

func TestScoreAmountCentsBoundariesAndDecisions(t *testing.T) {
	tests := []struct {
		cents    int64
		score    int
		decision Decision
		reasons  int
	}{
		{0, 0, Allow, 0}, {500000, 0, Allow, 0}, {500001, 40, Review, 1},
		{1000000, 40, Review, 1}, {1000001, 70, Deny, 2},
	}
	for _, tt := range tests {
		score, decision, reasons := ScoreAmountCents(tt.cents)
		if score != tt.score || decision != tt.decision || len(reasons) != tt.reasons {
			t.Errorf("%d cents: got %d %s %v", tt.cents, score, decision, reasons)
		}
		if score < 0 || score > 100 {
			t.Errorf("score out of range: %d", score)
		}
	}
}
