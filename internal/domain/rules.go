package domain

import "errors"

// RuleConfig contains synthetic, weighted policy thresholds; all limits are explicit.
type RuleConfig struct {
	ReviewAt, DenyAt                                        int
	HighAmountCents, VeryHighAmountCents                    int64
	HighAmountWeight, VeryHighAmountWeight                  int
	VelocityLimit                                           int
	VelocityWeight, GeoMismatchWeight, DeviceMismatchWeight int
	NewPayeeWeight, StructuringWeight                       int
	StructuringBandCents                                    int64
}

func DefaultRuleConfig() RuleConfig {
	return RuleConfig{ReviewAt: 40, DenyAt: 70, HighAmountCents: 500000, VeryHighAmountCents: 1000000, HighAmountWeight: 40, VeryHighAmountWeight: 30, VelocityLimit: 5, VelocityWeight: 25, GeoMismatchWeight: 20, DeviceMismatchWeight: 15, NewPayeeWeight: 15, StructuringWeight: 20, StructuringBandCents: 100000}
}
func (c RuleConfig) Validate() error {
	if c.ReviewAt < 1 || c.DenyAt < c.ReviewAt || c.DenyAt > 100 || c.HighAmountCents < 0 || c.VeryHighAmountCents < c.HighAmountCents || c.VelocityLimit < 1 || c.StructuringBandCents < 1 {
		return errors.New("invalid fraud rule thresholds")
	}
	for _, w := range []int{c.HighAmountWeight, c.VeryHighAmountWeight, c.VelocityWeight, c.GeoMismatchWeight, c.DeviceMismatchWeight, c.NewPayeeWeight, c.StructuringWeight} {
		if w < 0 || w > 100 {
			return errors.New("rule weight must be between 0 and 100")
		}
	}
	return nil
}

type Signals struct {
	AmountCents                           int64
	RecentCount                           int
	GeoMismatch, DeviceMismatch, NewPayee bool
	Structuring                           bool
}

func ScoreWithRules(s Signals, c RuleConfig) (int, Decision, []string, error) {
	if err := c.Validate(); err != nil {
		return 0, "", nil, err
	}
	if s.AmountCents < 0 || s.RecentCount < 0 {
		return 0, "", nil, errors.New("negative amount or velocity")
	}
	score := 0
	reasons := make([]string, 0, 7)
	add := func(hit bool, w int, reason string) {
		if hit {
			score += w
			reasons = append(reasons, reason)
		}
	}
	add(s.AmountCents > c.HighAmountCents, c.HighAmountWeight, "HIGH_AMOUNT")
	add(s.AmountCents > c.VeryHighAmountCents, c.VeryHighAmountWeight, "VERY_HIGH_AMOUNT")
	add(s.RecentCount >= c.VelocityLimit, c.VelocityWeight, "ACCOUNT_VELOCITY")
	add(s.GeoMismatch, c.GeoMismatchWeight, "GEO_MISMATCH")
	add(s.DeviceMismatch, c.DeviceMismatchWeight, "DEVICE_MISMATCH")
	add(s.NewPayee, c.NewPayeeWeight, "NEW_PAYEE")
	add(s.Structuring && s.AmountCents <= c.StructuringBandCents, c.StructuringWeight, "STRUCTURING")
	if score > 100 {
		score = 100
	}
	decision := Allow
	if score >= c.DenyAt {
		decision = Deny
	} else if score >= c.ReviewAt {
		decision = Review
	}
	return score, decision, reasons, nil
}
