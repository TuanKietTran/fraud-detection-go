package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/api"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OpenStore does not log connection information. Startup is bounded by caller.
func OpenStore(ctx context.Context, c Config) (repository.EventStore, func(), error) {
	if err := c.InitialRules.Validate(); err != nil {
		return nil, nil, ErrStartupRules
	}
	if c.DatabaseURL == "" {
		return repository.NewMemoryEvents(c.InitialRules), func() {}, nil
	}
	config, err := pgxpool.ParseConfig(c.DatabaseURL)
	if err != nil {
		return nil, nil, errors.New("invalid database configuration")
	}
	config.MaxConns = 4
	config.ConnConfig.ConnectTimeout = c.StartupTimeout
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, nil, errors.New("database pool initialization failed")
	}
	store, err := repository.NewPostgresEvents(ctx, pool, c.InitialRules)
	if err != nil {
		pool.Close()
		return nil, nil, errors.New("database event store initialization failed")
	}
	// Validate the active durable state as well as the incoming seed. The
	// constructor inserts only when absent; never merge/reset existing policy.
	if err := store.Update(ctx, func(*repository.EventState) error { return nil }); err != nil {
		pool.Close()
		return nil, nil, errors.New("database active policy validation failed")
	}
	return store, pool.Close, nil
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}
func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(b)
}
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func NewHTTPServer(c Config, store repository.EventStore, logger *slog.Logger) *http.Server {
	score, _ := query.NewScoreEventHandlerWithStore(store, time.Now, time.Minute)
	var requests, failures atomic.Uint64
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Request-ID", uuid.NewString())
			requests.Add(1)
			start := time.Now()
			wrapped := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(wrapped, req)
			status := wrapped.status
			if status == 0 {
				status = 200
			}
			if status >= 500 {
				failures.Add(1)
			}
			logger.Info("http request", "method", req.Method, "status", status, "duration", time.Since(start), "request_id", w.Header().Get("X-Request-ID"))
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if recover() != nil {
					logger.Error("http panic recovered", "request_id", w.Header().Get("X-Request-ID"))
					operationalError(w, 500)
				}
			}()
			next.ServeHTTP(w, req)
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			id := w.Header().Get("X-Request-ID")
			w.Header().Set("Content-Type", "application/json")
			body := fmt.Sprintf(`{"error":{"code":"request_timeout","message":"request deadline exceeded"},"request_id":%q}`, id)
			bufferedNext := http.HandlerFunc(func(buffer http.ResponseWriter, request *http.Request) {
				buffer.Header().Set("X-Request-ID", id)
				next.ServeHTTP(buffer, request)
			})
			http.TimeoutHandler(bufferedNext, c.RequestTimeout, body).ServeHTTP(w, req)
		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) { operationalError(w, 404) })
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) { operationalError(w, 405) })
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), c.RequestTimeout)
		defer cancel()
		if _, err := score.LoadRules(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(503)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": "not_ready", "message": "store unavailable"}, "request_id": w.Header().Get("X-Request-ID")})
			return
		}
		w.WriteHeader(200)
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]uint64{"requests_total": requests.Load(), "failures_total": failures.Load()})
	})
	r.Mount("/v1", api.NewHandlers(score).Routes())
	return &http.Server{Addr: c.Address, Handler: r, ReadHeaderTimeout: 2 * time.Second, ReadTimeout: c.RequestTimeout + 2*time.Second, WriteTimeout: c.RequestTimeout + 2*time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
}

func operationalError(w http.ResponseWriter, status int) {
	code := "not_found"
	if status == 405 {
		code = "method_not_allowed"
	}
	if status == 500 {
		code = "internal_error"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": http.StatusText(status)}, "request_id": w.Header().Get("X-Request-ID")})
}

// Serve accepts an injected listener so tests synchronize on readiness without sleeps.
func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		stop, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := server.Shutdown(stop)
		if err != nil {
			_ = server.Close()
		}
		serveErr := <-done
		if err != nil {
			return err
		}
		if !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}

func Run(ctx context.Context, c Config, logger *slog.Logger) error {
	startup, cancel := context.WithTimeout(ctx, c.StartupTimeout)
	store, closeStore, err := OpenStore(startup, c)
	cancel()
	if err != nil {
		return err
	}
	defer closeStore()
	server := NewHTTPServer(c, store, logger)
	listener, err := net.Listen("tcp", c.Address)
	if err != nil {
		return err
	}
	logger.Info("fraud server listening", "address", listener.Addr().String(), "durable", c.DatabaseURL != "", "policy_mode", "seed-only; existing durable policy retained")
	return Serve(ctx, server, listener, c.ShutdownTimeout)
}
