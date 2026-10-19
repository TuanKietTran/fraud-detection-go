package playback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
	"github.com/TuanKietTran/fraud-detection-go/internal/service"
)

// TestR1ProductionRouterTenThousand exercises the production router and memory
// event store through a loopback HTTP listener, including deterministic faults.
func TestR1ProductionRouterTenThousand(t *testing.T) {
	cfg, err := service.LoadConfig(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	store := repository.NewMemoryEvents(domain.DefaultRuleConfig())
	production := service.NewHTTPServer(cfg, store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	var mu sync.Mutex
	attempts := make(map[string]int)
	var injected503, injected429, injectedTimeout atomic.Int64
	faults := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(data))
		var body struct {
			EventID string `json:"event_id"`
		}
		_ = json.Unmarshal(data, &body)
		mu.Lock()
		attempts[body.EventID]++
		n := attempts[body.EventID]
		mu.Unlock()
		var index int
		_, _ = fmt.Sscanf(body.EventID, "r1-%d", &index)
		if n == 1 && index%19 == 0 && index%98 != 1 {
			if index%100 == 0 {
				injected429.Add(1)
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				injected503.Add(1)
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			return
		}
		if n == 1 && index%98 == 1 {
			injectedTimeout.Add(1)
			time.Sleep(2100 * time.Millisecond)
			if r.Context().Err() != nil {
				return
			}
		}
		production.Handler.ServeHTTP(w, r)
	})
	server := httptest.NewServer(faults)
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 0
	client.Transport = &http.Transport{MaxConnsPerHost: 32, MaxIdleConnsPerHost: 32}
	policy := DefaultPolicy()
	policy.Concurrency = 32
	calls := make([]Call, TargetRequestsPerMinute)
	for i := range calls {
		event := i
		if i >= 9800 {
			event = (i - 9800) * 49
		}
		calls[i] = Call{Method: http.MethodPost, Path: "/v1/score", Body: []byte(fmt.Sprintf(`{"event_id":"r1-%d","amount":"10.00","currency":"USD","account":"acct-%d"}`, event, event))}
	}
	ctx, cancel := context.WithTimeout(context.Background(), Window)
	defer cancel()
	start := time.Now()
	outcomes, err := Play(ctx, client, server.URL, calls, policy)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed >= Window {
		t.Fatalf("10,000 calls took %s, deadline %s", elapsed, Window)
	}
	for i, outcome := range outcomes {
		if outcome.Err != nil || outcome.Status != http.StatusOK {
			t.Fatalf("call %d: status=%d attempts=%d err=%v body=%s", i, outcome.Status, outcome.Attempts, outcome.Err, outcome.Body)
		}
	}
	var eventCount int
	if err := store.Update(context.Background(), func(state *repository.EventState) error { eventCount = len(state.Events); return nil }); err != nil {
		t.Fatal(err)
	}
	if eventCount != 9800 {
		t.Fatalf("stored decisions=%d, want 9800 unique event identities", eventCount)
	}
	var saw429, sawTimeout bool
	for id, n := range attempts {
		var index int
		_, _ = fmt.Sscanf(id, "r1-%d", &index)
		if index%100 == 0 && n >= 2 {
			saw429 = true
		}
		if index%98 == 1 && n >= 2 {
			sawTimeout = true
		}
	}
	if !saw429 || !sawTimeout {
		t.Fatalf("injected 429=%v timeout retry=%v", saw429, sawTimeout)
	}
	if injected503.Load()+injected429.Load() < TargetRequestsPerMinute/20 || injected429.Load() == 0 || injectedTimeout.Load() < TargetRequestsPerMinute/100 {
		t.Fatalf("faults 503=%d 429=%d timeout=%d", injected503.Load(), injected429.Load(), injectedTimeout.Load())
	}
	t.Logf("production-router replay: %d logical calls, %d unique decisions, elapsed=%s", len(outcomes), eventCount, elapsed)
}
