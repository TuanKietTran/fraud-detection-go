package repository

import (
	"context"
	"sync"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

// EventInput is the canonical replay payload. RecentCount is a legacy signal;
// the service combines it with, but never substitutes it for, its own count.
type EventInput struct {
	EventID                                            string
	AmountCents                                        int64
	Account                                            string
	RecentCount                                        int
	GeoMismatch, DeviceMismatch, NewPayee, Structuring bool
}

type EventRecord struct {
	Input       EventInput
	Result      domain.RiskDecision
	RecordedAt  time.Time
	RecentCount int
	Rules       domain.RuleConfig
}

type EventState struct {
	Events   map[string]EventRecord
	Feedback map[string]string
	Rules    domain.RuleConfig
}

// EventStore serializes a complete decision/velocity/feedback transaction.
// A callback error rolls back all changes. Callbacks must not retain state.
type EventStore interface {
	Update(context.Context, func(*EventState) error) error
}

type MemoryEvents struct {
	mu    sync.Mutex
	state EventState
}

func NewMemoryEvents(config domain.RuleConfig) *MemoryEvents {
	return &MemoryEvents{state: EventState{Events: make(map[string]EventRecord), Feedback: make(map[string]string), Rules: config}}
}

func cloneState(s EventState) EventState {
	copy := EventState{Events: make(map[string]EventRecord, len(s.Events)), Feedback: make(map[string]string, len(s.Feedback)), Rules: s.Rules}
	for id, event := range s.Events {
		event.Result.ReasonCodes = append([]string{}, event.Result.ReasonCodes...)
		copy.Events[id] = event
	}
	for id, label := range s.Feedback {
		copy.Feedback[id] = label
	}
	return copy
}

func (m *MemoryEvents) Update(ctx context.Context, fn func(*EventState) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	next := cloneState(m.state)
	if err := fn(&next); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	m.state = cloneState(next)
	return nil
}
