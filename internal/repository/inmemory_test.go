package repository

import (
	"context"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/google/uuid"
)

func TestLegacyDecisionRepositoryRoundTrip(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryFraud()
	absent, err := store.GetDecision(ctx, "missing")
	if err != nil || absent != nil {
		t.Fatalf("absent: %+v %v", absent, err)
	}
	input := domain.RiskDecision{EventID: "e", Score: 40, Decision: domain.Review}
	if err := store.SaveDecision(ctx, input); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetDecision(ctx, "e")
	if err != nil || got == nil || got.ID == uuid.Nil || got.Score != 40 {
		t.Fatalf("roundtrip: %+v %v", got, err)
	}
	id := uuid.New()
	input.ID = id
	input.Score = 70
	if err := store.SaveDecision(ctx, input); err != nil {
		t.Fatal(err)
	}
	got, err = store.GetDecision(ctx, "e")
	if err != nil || got.ID != id || got.Score != 70 {
		t.Fatalf("explicit id update: %+v %v", got, err)
	}
}
