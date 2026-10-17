package query

import (
	"context"
	"testing"
)

func TestExplanationRulesAndFeedback(t *testing.T) {
	h := NewScoreEventHandler()
	q := ScoreEventQuery{EventID: "e", Account: "a", AmountCents: 700000}
	result, err := h.Handle(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if explained, ok := h.Explain("e"); !ok || explained.Score != result.Score {
		t.Fatal("explanation missing")
	}
	if _, ok := h.Explain("absent"); ok {
		t.Fatal("unexpected explanation")
	}
	if h.Rules().ReviewAt != 40 {
		t.Fatal("wrong initial config")
	}
	if err = h.Feedback("absent", "CONFIRMED_FRAUD"); err == nil {
		t.Fatal("feedback for absent event accepted")
	}
	if err = h.Feedback("e", "wrong"); err == nil {
		t.Fatal("invalid label accepted")
	}
	if err = h.Feedback("e", "CONFIRMED_FRAUD"); err != nil {
		t.Fatal(err)
	}
	if err = h.Feedback("e", "CONFIRMED_FRAUD"); err != nil {
		t.Fatal(err)
	}
	if h.Rules().ReviewAt != 39 {
		t.Fatal("confirmed fraud did not adjust threshold deterministically")
	}
	if err = h.Feedback("e", "FALSE_POSITIVE"); err == nil {
		t.Fatal("conflicting feedback accepted")
	}
}
