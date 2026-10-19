package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
)

type panicStore struct{}

func (panicStore) Update(context.Context, func(*repository.EventState) error) error {
	panic("synthetic sensitive panic")
}

func TestOperationalPanicIsContained(t *testing.T) {
	server := NewHTTPServer(defaults(), panicStore{}, quiet())
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || rec.Code != 500 || body.Error.Code != "internal_error" || body.RequestID != rec.Header().Get("X-Request-ID") {
		t.Fatalf("%d %s %v", rec.Code, rec.Body.String(), err)
	}
}

func TestMetricsExposeRuntimeResourcesAndPprofIsLoopbackOnly(t *testing.T) {
	server := NewHTTPServer(defaults(), repository.NewMemoryEvents(domain.DefaultRuleConfig()), quiet())
	metrics := httptest.NewRecorder()
	server.Handler.ServeHTTP(metrics, httptest.NewRequest("GET", "/metrics", nil))
	var values map[string]uint64
	if err := json.Unmarshal(metrics.Body.Bytes(), &values); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"goroutines", "heap_alloc_bytes", "heap_sys_bytes", "gc_count", "gc_pause_total_ns", "connections"} {
		if _, ok := values[key]; !ok {
			t.Errorf("missing runtime metric %q", key)
		}
	}
	remote := httptest.NewRequest("GET", "/debug/pprof/", nil)
	remote.RemoteAddr = "203.0.113.1:1234"
	blocked := httptest.NewRecorder()
	server.Handler.ServeHTTP(blocked, remote)
	if blocked.Code != 404 {
		t.Fatalf("non-loopback pprof status=%d", blocked.Code)
	}
	local := httptest.NewRequest("GET", "/debug/pprof/", nil)
	local.RemoteAddr = "127.0.0.1:1234"
	allowed := httptest.NewRecorder()
	server.Handler.ServeHTTP(allowed, local)
	if allowed.Code != 200 {
		t.Fatalf("loopback pprof status=%d", allowed.Code)
	}
}

func TestOperationalFallbackErrorsAreJSON(t *testing.T) {
	server := NewHTTPServer(defaults(), repository.NewMemoryEvents(domain.DefaultRuleConfig()), quiet())
	for _, tc := range []struct {
		method, path string
		want         int
	}{{"GET", "/unknown", 404}, {"POST", "/healthz", 405}} {
		rec := httptest.NewRecorder()
		server.Handler.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || rec.Code != tc.want || body["request_id"] != rec.Header().Get("X-Request-ID") {
			t.Fatalf("%d %s %v", rec.Code, rec.Body.String(), err)
		}
	}
}
