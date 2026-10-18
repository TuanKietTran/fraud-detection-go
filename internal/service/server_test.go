package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
	"gopkg.in/yaml.v3"
)

func defaults() Config    { c, _ := LoadConfig(func(string) string { return "" }); return c }
func quiet() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }
func TestConfiguration(t *testing.T) {
	c := defaults()
	if c.Address != ":8103" || c.RequestTimeout != 5*time.Second {
		t.Fatal(c)
	}
	for _, tc := range []struct{ key, value string }{
		{"HTTP_ADDR", "bad"}, {"HTTP_ADDR", ":0"}, {"HTTP_ADDR", ":65536"}, {"HTTP_ADDR", ":not-a-port"},
		{"DATABASE_URL", "https://localhost/db"}, {"DATABASE_URL", "postgres:///db"}, {"DATABASE_URL", "%"},
		{"REQUEST_TIMEOUT", "0s"}, {"STARTUP_TIMEOUT", "-1s"}, {"SHUTDOWN_TIMEOUT", "2m"}, {"REQUEST_TIMEOUT", "bad"},
	} {
		if _, err := LoadConfig(func(key string) string {
			if key == tc.key {
				return tc.value
			}
			return ""
		}); err == nil {
			t.Errorf("accepted %s", tc.key)
		}
	}
	c, err := LoadConfig(func(key string) string {
		return map[string]string{"HTTP_ADDR": "127.0.0.1:8103", "REQUEST_TIMEOUT": "1s", "DATABASE_URL": "postgres://localhost/test"}[key]
	})
	if err != nil || c.RequestTimeout != time.Second {
		t.Fatalf("%+v %v", c, err)
	}
}

type unavailable struct{}

func (unavailable) Update(context.Context, func(*repository.EventState) error) error {
	return errors.New("synthetic failure")
}

type blockedStore struct {
	started chan struct{}
	done    chan struct{}
}

func (s blockedStore) Update(ctx context.Context, fn func(*repository.EventState) error) error {
	close(s.started)
	<-ctx.Done()
	close(s.done)
	return ctx.Err()
}

func TestOperationalRoutesMetricsAndUnavailableReadiness(t *testing.T) {
	spec, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Responses map[string]any `yaml:"responses"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		responses := doc.Paths[path]["get"].Responses
		if len(responses) != 3 || responses["200"] == nil || responses["500"] == nil || responses["503"] == nil {
			t.Fatalf("operational contract drift %s: %v", path, responses)
		}
	}
	for _, tc := range []struct {
		store repository.EventStore
		ready int
	}{{repository.NewMemoryEvents(domain.DefaultRuleConfig()), 200}, {unavailable{}, 503}} {
		server := NewHTTPServer(defaults(), tc.store, quiet())
		for _, route := range []struct {
			path string
			want int
		}{{"/healthz", 200}, {"/readyz", tc.ready}, {"/metrics", 200}} {
			rec := httptest.NewRecorder()
			server.Handler.ServeHTTP(rec, httptest.NewRequest("GET", route.path, nil))
			if rec.Code != route.want || rec.Header().Get("X-Request-ID") == "" {
				t.Fatalf("%s: %d", route.path, rec.Code)
			}
			if route.path == "/metrics" {
				var data map[string]uint64
				if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil || data["requests_total"] != 3 {
					t.Fatalf("metrics %s %v", rec.Body.String(), err)
				}
			}
		}
		if server.ReadHeaderTimeout <= 0 || server.WriteTimeout <= 0 || server.IdleTimeout <= 0 || server.MaxHeaderBytes == 0 {
			t.Fatal("unbounded server")
		}
	}
}

func TestRequestTimeoutCancelsStoreAndReturnsJSON(t *testing.T) {
	c := defaults()
	c.RequestTimeout = time.Millisecond
	store := blockedStore{started: make(chan struct{}), done: make(chan struct{})}
	server := NewHTTPServer(c, store, quiet())
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, httptest.NewRequest("GET", "/v1/rules", nil))
	<-store.started
	<-store.done
	if rec.Code != 503 && rec.Code != 408 {
		t.Fatalf("timeout status %d", rec.Code)
	}
	var body struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.RequestID == "" || body.RequestID != rec.Header().Get("X-Request-ID") || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("timeout: %v %s %v", err, rec.Body.String(), rec.Header())
	}
}

func TestServeRealListenerGracefulCancellationAndServeFailure(t *testing.T) {
	store, closeStore, err := OpenStore(context.Background(), defaults())
	if err != nil {
		t.Fatal(err)
	}
	defer closeStore()
	server := NewHTTPServer(defaults(), store, quiet())
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, server, listener, time.Second) }()
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal(response.StatusCode)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
	if err := Serve(context.Background(), NewHTTPServer(defaults(), store, quiet()), listener, time.Second); err == nil {
		t.Fatal("closed listener accepted")
	}
}

func TestRunStartupAndListenFailuresAreBoundedAndSanitized(t *testing.T) {
	c := defaults()
	c.DatabaseURL = "not a postgres url"
	if _, _, err := OpenStore(context.Background(), c); err == nil {
		t.Fatal("bad database configuration")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	c = defaults()
	c.Address = listener.Addr().String()
	if err := Run(context.Background(), c, quiet()); err == nil {
		t.Fatal("occupied listen port accepted")
	}
	c.DatabaseURL = "postgres://test-password-not-a-secret@127.0.0.1:1/test?sslmode=disable"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := OpenStore(ctx, c); err == nil || strings.Contains(err.Error(), "test-password") {
		t.Fatalf("unsafe startup error %v", err)
	}
}
