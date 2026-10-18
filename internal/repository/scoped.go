package repository

import (
	"context"
	"errors"
)

// EventScope limits the transactional view to the addressed event/feedback and,
// when History is true, every event for Account. Rules are always included.
// Unknown/colliding writes outside this scope fail atomically, never disappear.
type EventScope struct {
	EventID, Account string
	History          bool
}

type ScopedEventStore interface {
	UpdateScoped(context.Context, EventScope, func(*EventState) error) error
}

func (m *MemoryEvents) index(id, account string) {
	if m.byAccount[account] == nil {
		m.byAccount[account] = make(map[string]struct{})
	}
	m.byAccount[account][id] = struct{}{}
}
func (m *MemoryEvents) unindex(id, account string) {
	delete(m.byAccount[account], id)
	if len(m.byAccount[account]) == 0 {
		delete(m.byAccount, account)
	}
}

// UpdateScoped retains the full Update rollback/copy guarantees but avoids
// copying unrelated accounts. Per-account history is still unbounded.
func (m *MemoryEvents) UpdateScoped(ctx context.Context, scope EventScope, fn func(*EventState) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	selected := make(map[string]EventRecord)
	if scope.History {
		for id := range m.byAccount[scope.Account] {
			selected[id] = m.state.Events[id]
		}
	}
	if scope.EventID != "" {
		if event, ok := m.state.Events[scope.EventID]; ok {
			selected[scope.EventID] = event
		}
	}
	initial := EventState{Events: selected, Feedback: make(map[string]string), Rules: m.state.Rules}
	if label, ok := m.state.Feedback[scope.EventID]; ok && scope.EventID != "" {
		initial.Feedback[scope.EventID] = label
	}
	next := cloneState(initial)
	if err := fn(&next); err != nil {
		return err
	}
	if err := next.Rules.Validate(); err != nil {
		return err
	}
	for id, event := range next.Events {
		allowed := (scope.EventID != "" && id == scope.EventID) || (scope.History && event.Input.Account == scope.Account)
		_, previouslySelected := selected[id]
		_, exists := m.state.Events[id]
		if !allowed || (exists && !previouslySelected) {
			return errors.New("event write outside transaction scope")
		}
	}
	for id := range next.Feedback {
		if scope.EventID == "" || id != scope.EventID {
			return errors.New("feedback write outside transaction scope")
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Allocate defensive copies before changing committed state.
	committed := cloneState(next)
	for id, old := range selected {
		delete(m.state.Events, id)
		m.unindex(id, old.Input.Account)
	}
	for id, event := range committed.Events {
		m.state.Events[id] = event
		m.index(id, event.Input.Account)
	}
	if scope.EventID != "" {
		delete(m.state.Feedback, scope.EventID)
		if label, ok := committed.Feedback[scope.EventID]; ok {
			m.state.Feedback[scope.EventID] = label
		}
	}
	m.state.Rules = committed.Rules
	return nil
}
