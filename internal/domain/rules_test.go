package domain

import "testing"

func TestWeightedRulesAndValidation(t *testing.T) {
	c := DefaultRuleConfig()
	score, d, reasons, err := ScoreWithRules(Signals{AmountCents: 500001, RecentCount: 5, GeoMismatch: true, DeviceMismatch: true, NewPayee: true, Structuring: true}, c)
	if err != nil || score != 100 || d != Deny || len(reasons) != 5 {
		t.Fatalf("got %d %s %v %v", score, d, reasons, err)
	}
	score, d, reasons, err = ScoreWithRules(Signals{AmountCents: 10000, Structuring: true}, DefaultRuleConfig())
	if err != nil || score != 20 || d != Allow || len(reasons) != 1 {
		t.Fatalf("structuring got %d %s %v %v", score, d, reasons, err)
	}
	c.VelocityWeight = -1
	if c.Validate() == nil {
		t.Fatal("negative weight accepted")
	}
	if _, _, _, err := ScoreWithRules(Signals{AmountCents: -1}, DefaultRuleConfig()); err == nil {
		t.Fatal("negative amount accepted")
	}
}
