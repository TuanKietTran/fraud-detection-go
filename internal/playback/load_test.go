package playback

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/api"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/go-chi/chi/v5"
)

func failFirst(next http.Handler) http.Handler {
	var mu sync.Mutex
	seen := map[string]int{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		mu.Lock()
		seen[string(body)]++
		n := seen[string(body)]
		mu.Unlock()
		if n == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TestThousandPerMinuteAndPlayback(t *testing.T) {
	h := api.NewHandlers(query.NewScoreEventHandler())
	router := chi.NewRouter()
	router.Use(failFirst)
	router.Mount("/v1", h.Routes())
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	amounts := []int{100, 6000, 20000}
	calls := make([]Call, RequestsPerMinute)
	for i := range calls {
		calls[i] = Call{
			Method: http.MethodPost,
			Path:   "/v1/score",
			Body:   []byte(fmt.Sprintf(`{"event_id":"e-%d","amount":%d,"account":"a"}`, i, amounts[i%len(amounts)])),
		}
	}
	start := time.Now()
	first, err := Play(context.Background(), srv.Client(), srv.URL, calls, DefaultPolicy())
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed >= Window {
		t.Fatalf("handled %d requests in %s, want under %s", len(calls), elapsed, Window)
	}
	t.Logf("first play %d requests in %s", len(calls), elapsed)
	var sawAllow, sawReview, sawBlock bool
	for i, o := range first {
		if o.Err != nil || o.Status != http.StatusOK || o.Attempts != 2 {
			t.Fatalf("call %d status %d attempts %d err %v body %s", i, o.Status, o.Attempts, o.Err, o.Body)
		}
		body := string(o.Body)
		sawAllow = sawAllow || containsDecision(body, "ALLOW")
		sawReview = sawReview || containsDecision(body, "REVIEW")
		sawBlock = sawBlock || containsDecision(body, "BLOCK")
	}
	if !sawAllow || !sawReview || !sawBlock {
		t.Fatalf("decisions allow=%v review=%v block=%v", sawAllow, sawReview, sawBlock)
	}

	second, err := Play(context.Background(), srv.Client(), srv.URL, calls, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	for i, o := range second {
		if o.Attempts != 1 || o.Status != http.StatusOK || string(o.Body) != string(first[i].Body) {
			t.Fatalf("playback %d status %d body %s want %s", i, o.Status, o.Body, first[i].Body)
		}
	}
}

func containsDecision(body, decision string) bool {
	return len(body) > 0 && (body == decision || len(body) >= len(decision) && search(body, `"Decision":"`+decision+`"`))
}

func search(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
