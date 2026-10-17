package repository

import (
	"context"
	"sync"

	"github.com/example/fraud-detection/internal/domain"
	"github.com/google/uuid"
)

type InMemoryFraud struct {
	mu        sync.RWMutex
	decisions map[string]domain.RiskDecision
}

func NewInMemoryFraud() *InMemoryFraud {
	return &InMemoryFraud{
		decisions: make(map[string]domain.RiskDecision),
	}
}

func (r *InMemoryFraud) SaveDecision(ctx context.Context, d domain.RiskDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	r.decisions[d.EventID] = d
	return nil
}

func (r *InMemoryFraud) GetDecision(ctx context.Context, eventID string) (*domain.RiskDecision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.decisions[eventID]
	if !ok {
		return nil, nil
	}
	return &d, nil
}
