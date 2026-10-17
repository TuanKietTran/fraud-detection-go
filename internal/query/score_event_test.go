package query

import (
	"context"
	"sync"
	"testing"
)

func TestReplayReturnsSameAndRejectsChangedPayload(t *testing.T) {
	h := NewScoreEventHandler()
	q := ScoreEventQuery{EventID: "evt", Account: "a", AmountCents: 600001}
	first, err := h.Handle(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	again, err := h.Handle(context.Background(), q)
	if err != nil || first.Score != again.Score || first.Decision != again.Decision {
		t.Fatalf("replay mismatch: %#v %#v %v", first, again, err)
	}
	q.AmountCents++
	if _, err = h.Handle(context.Background(), q); err == nil {
		t.Fatal("changed payload accepted for event_id")
	}
}
func TestConcurrentReplayIsRaceSafe(t *testing.T) {
	h := NewScoreEventHandler()
	q := ScoreEventQuery{EventID: "evt", Account: "a", AmountCents: 42}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.Handle(context.Background(), q); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
