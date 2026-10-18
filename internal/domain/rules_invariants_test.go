package domain

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func assertRejectedPolicy(t *testing.T, c RuleConfig) {
	t.Helper()
	if c.Validate() == nil {
		t.Fatal("invalid policy accepted")
	}
	score, decision, reasons, err := ScoreWithRules(Signals{}, c)
	if err == nil || score != 0 || decision != "" || reasons != nil {
		t.Fatalf("invalid policy leaked decision: %d %s %v %v", score, decision, reasons, err)
	}
}

func TestEveryRuleConfigThresholdAndRelationshipInvariant(t *testing.T) {
	invalid := []struct {
		name   string
		change func(*RuleConfig)
	}{
		{"review-zero", func(c *RuleConfig) { c.ReviewAt = 0 }},
		{"review-negative", func(c *RuleConfig) { c.ReviewAt = -1 }},
		{"review-above-range", func(c *RuleConfig) { c.ReviewAt = 101 }},
		{"deny-before-review", func(c *RuleConfig) { c.DenyAt = c.ReviewAt - 1 }},
		{"deny-zero", func(c *RuleConfig) { c.DenyAt = 0 }},
		{"deny-above-range", func(c *RuleConfig) { c.DenyAt = 101 }},
		{"negative-high-amount", func(c *RuleConfig) { c.HighAmountCents = -1 }},
		{"very-high-before-high", func(c *RuleConfig) { c.VeryHighAmountCents = c.HighAmountCents - 1 }},
		{"negative-very-high", func(c *RuleConfig) { c.VeryHighAmountCents = -1 }},
		{"velocity-zero", func(c *RuleConfig) { c.VelocityLimit = 0 }},
		{"velocity-negative", func(c *RuleConfig) { c.VelocityLimit = -1 }},
		{"structuring-band-zero", func(c *RuleConfig) { c.StructuringBandCents = 0 }},
		{"structuring-band-negative", func(c *RuleConfig) { c.StructuringBandCents = -1 }},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) { c := DefaultRuleConfig(); tc.change(&c); assertRejectedPolicy(t, c) })
	}
	valid := []struct {
		name   string
		change func(*RuleConfig)
	}{
		{"review-min", func(c *RuleConfig) { c.ReviewAt = 1 }},
		{"equal-decision-limits", func(c *RuleConfig) { c.ReviewAt = c.DenyAt }},
		{"decision-min", func(c *RuleConfig) { c.ReviewAt = 1; c.DenyAt = 1 }},
		{"decision-max", func(c *RuleConfig) { c.ReviewAt = 100; c.DenyAt = 100 }},
		{"zero-amount-limits", func(c *RuleConfig) { c.HighAmountCents = 0; c.VeryHighAmountCents = 0 }},
		{"equal-amount-limits", func(c *RuleConfig) { c.VeryHighAmountCents = c.HighAmountCents }},
		{"int64-amount-limits", func(c *RuleConfig) { c.HighAmountCents = math.MaxInt64; c.VeryHighAmountCents = math.MaxInt64 }},
		{"velocity-min", func(c *RuleConfig) { c.VelocityLimit = 1 }},
		{"structuring-band-min", func(c *RuleConfig) { c.StructuringBandCents = 1 }},
		{"structuring-band-max", func(c *RuleConfig) { c.StructuringBandCents = math.MaxInt64 }},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultRuleConfig()
			tc.change(&c)
			if err := c.Validate(); err != nil {
				t.Fatalf("valid boundary rejected: %v", err)
			}
		})
	}
}

func TestEveryRuleWeightRejectsOutOfRangeAndAcceptsEndpoints(t *testing.T) {
	for _, weight := range []struct {
		name  string
		field func(*RuleConfig) *int
	}{
		{"high", func(c *RuleConfig) *int { return &c.HighAmountWeight }},
		{"very-high", func(c *RuleConfig) *int { return &c.VeryHighAmountWeight }},
		{"velocity", func(c *RuleConfig) *int { return &c.VelocityWeight }},
		{"geo", func(c *RuleConfig) *int { return &c.GeoMismatchWeight }},
		{"device", func(c *RuleConfig) *int { return &c.DeviceMismatchWeight }},
		{"new-payee", func(c *RuleConfig) *int { return &c.NewPayeeWeight }},
		{"structuring", func(c *RuleConfig) *int { return &c.StructuringWeight }},
	} {
		for _, value := range []int{-1, 101} {
			t.Run(fmt.Sprintf("%s-reject-%d", weight.name, value), func(t *testing.T) { c := DefaultRuleConfig(); *weight.field(&c) = value; assertRejectedPolicy(t, c) })
		}
		for _, value := range []int{0, 100} {
			t.Run(fmt.Sprintf("%s-accept-%d", weight.name, value), func(t *testing.T) {
				c := DefaultRuleConfig()
				*weight.field(&c) = value
				if err := c.Validate(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestScoringRejectsNegativeSignalsWithoutPartialDecision(t *testing.T) {
	for _, signals := range []Signals{{AmountCents: -1}, {RecentCount: -1}, {AmountCents: -1, RecentCount: -1}} {
		score, decision, reasons, err := ScoreWithRules(signals, DefaultRuleConfig())
		if err == nil || score != 0 || decision != "" || reasons != nil {
			t.Fatalf("invalid signal leaked decision: %+v => %d %s %v %v", signals, score, decision, reasons, err)
		}
	}
}

func TestWeightedDecisionAndRuleActivationBoundaries(t *testing.T) {
	c := DefaultRuleConfig()
	for _, tc := range []struct {
		name     string
		signals  Signals
		score    int
		decision Decision
		reasons  []string
	}{
		{"zero", Signals{}, 0, Allow, []string{}},
		{"high-at", Signals{AmountCents: c.HighAmountCents}, 0, Allow, []string{}},
		{"high-over-review-at", Signals{AmountCents: c.HighAmountCents + 1}, 40, Review, []string{"HIGH_AMOUNT"}},
		{"very-high-at", Signals{AmountCents: c.VeryHighAmountCents}, 40, Review, []string{"HIGH_AMOUNT"}},
		{"very-high-over-deny-at", Signals{AmountCents: c.VeryHighAmountCents + 1}, 70, Deny, []string{"HIGH_AMOUNT", "VERY_HIGH_AMOUNT"}},
		{"velocity-before", Signals{RecentCount: c.VelocityLimit - 1}, 0, Allow, []string{}},
		{"velocity-at", Signals{RecentCount: c.VelocityLimit}, 25, Allow, []string{"ACCOUNT_VELOCITY"}},
		{"context-review", Signals{GeoMismatch: true, DeviceMismatch: true, NewPayee: true}, 50, Review, []string{"GEO_MISMATCH", "DEVICE_MISMATCH", "NEW_PAYEE"}},
		{"structuring-band-at", Signals{AmountCents: c.StructuringBandCents, Structuring: true}, 20, Allow, []string{"STRUCTURING"}},
		{"structuring-band-over", Signals{AmountCents: c.StructuringBandCents + 1, Structuring: true}, 0, Allow, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			score, decision, reasons, err := ScoreWithRules(tc.signals, c)
			if err != nil || score != tc.score || decision != tc.decision || !reflect.DeepEqual(reasons, tc.reasons) {
				t.Fatalf("got %d %s %v %v", score, decision, reasons, err)
			}
		})
	}
	c.ReviewAt = 25
	c.DenyAt = 50
	for _, tc := range []struct {
		signals  Signals
		score    int
		decision Decision
	}{
		{Signals{GeoMismatch: true}, 20, Allow},
		{Signals{RecentCount: 5}, 25, Review},
		{Signals{GeoMismatch: true, DeviceMismatch: true, NewPayee: true}, 50, Deny},
	} {
		score, decision, _, err := ScoreWithRules(tc.signals, c)
		if err != nil || score != tc.score || decision != tc.decision {
			t.Fatalf("custom threshold: %d %s %v", score, decision, err)
		}
	}
	c = DefaultRuleConfig()
	c.ReviewAt = 1
	c.DenyAt = 1
	c.HighAmountWeight = 1
	score, decision, _, err := ScoreWithRules(Signals{AmountCents: c.HighAmountCents + 1}, c)
	if err != nil || score != 1 || decision != Deny {
		t.Fatal("equal decision limits must prioritize deny", score, decision, err)
	}
}

func TestScoreCapRetainsAllReasonsAndDeterministicSignalProperties(t *testing.T) {
	c := DefaultRuleConfig()
	c.StructuringBandCents = c.VeryHighAmountCents + 1
	signals := Signals{AmountCents: c.VeryHighAmountCents + 1, RecentCount: c.VelocityLimit, GeoMismatch: true, DeviceMismatch: true, NewPayee: true, Structuring: true}
	score, decision, reasons, err := ScoreWithRules(signals, c)
	want := []string{"HIGH_AMOUNT", "VERY_HIGH_AMOUNT", "ACCOUNT_VELOCITY", "GEO_MISMATCH", "DEVICE_MISMATCH", "NEW_PAYEE", "STRUCTURING"}
	if err != nil || score != 100 || decision != Deny || !reflect.DeepEqual(reasons, want) {
		t.Fatalf("cap lost reasons: %d %s %v %v", score, decision, reasons, err)
	}
	c = DefaultRuleConfig()
	for mask := 0; mask < 16; mask++ {
		for _, amount := range []int64{0, c.StructuringBandCents, c.HighAmountCents, c.HighAmountCents + 1, c.VeryHighAmountCents, c.VeryHighAmountCents + 1, math.MaxInt64} {
			s := Signals{AmountCents: amount, RecentCount: c.VelocityLimit, GeoMismatch: mask&1 != 0, DeviceMismatch: mask&2 != 0, NewPayee: mask&4 != 0, Structuring: mask&8 != 0}
			first, d, why, err := ScoreWithRules(s, c)
			again, d2, why2, err2 := ScoreWithRules(s, c)
			if err != nil || err2 != nil || first < 0 || first > 100 || first != again || d != d2 || !reflect.DeepEqual(why, why2) {
				t.Fatal("non-deterministic/unbounded score", s)
			}
			seen := map[string]bool{}
			for _, reason := range why {
				if seen[reason] {
					t.Fatal("duplicate reason", reason)
				}
				seen[reason] = true
			}
			s.GeoMismatch = true
			raised, _, _, err := ScoreWithRules(s, c)
			if err != nil || raised < first {
				t.Fatal("adding a nonnegative geo signal reduced score")
			}
		}
	}
}
