package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
)

// Fault is after a successful commit but before a successful HTTP response.
type ambiguousStore struct {
	inner               *repository.MemoryEvents
	first               atomic.Bool
	committed, finished chan struct{}
}

func (s *ambiguousStore) Update(ctx context.Context, fn func(*repository.EventState) error) error {
	if err := s.inner.Update(ctx, fn); err != nil {
		return err
	}
	if !s.first.Swap(true) {
		close(s.committed)
		<-ctx.Done()
		close(s.finished)
		return ctx.Err()
	}
	return nil
}
func TestRetryAfterCommittedButCanceledHTTPResponse(t *testing.T) {
	memory := repository.NewMemoryEvents(domain.DefaultRuleConfig())
	fault := &ambiguousStore{inner: memory, committed: make(chan struct{}), finished: make(chan struct{})}
	server := NewHTTPServer(defaults(), fault, quiet())
	payload := `{"event_id":"ambiguous","account":"isolated","amount":"950.00","currency":"USD"}`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	first := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		server.Handler.ServeHTTP(first, httptest.NewRequest("POST", "/v1/score", strings.NewReader(payload)).WithContext(ctx))
	}()
	select {
	case <-fault.committed:
		cancel()
	case <-ctx.Done():
		t.Fatal("commit was not observed")
	}
	<-done
	<-fault.finished
	if first.Code != 408 && first.Code != 503 {
		t.Fatalf("want ambiguous timeout, got %d", first.Code)
	}
	replay := httptest.NewRecorder()
	server.Handler.ServeHTTP(replay, httptest.NewRequest("POST", "/v1/score", strings.NewReader(payload)))
	var result struct {
		EventID  string `json:"event_id"`
		Score    int    `json:"score"`
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal(replay.Body.Bytes(), &result); err != nil || replay.Code != 200 || result.EventID != "ambiguous" || result.Score != 0 || result.Decision != "ALLOW" {
		t.Fatalf("replay %d %s %v", replay.Code, replay.Body.String(), err)
	}
	if err := memory.Update(context.Background(), func(s *repository.EventState) error {
		if len(s.Events) != 1 || s.Events["ambiguous"].RecentCount != 0 || s.Events["ambiguous"].StructuringCount != 0 {
			t.Fatal("retry added another velocity/pattern effect")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
