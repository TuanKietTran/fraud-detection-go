package repository_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
	"github.com/TuanKietTran/fraud-detection-go/internal/service"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A failed database startup FAILS this test: a skip is not parity evidence.
func TestPostgresEventParity(t *testing.T) {
	dir := t.TempDir()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint32(listener.Addr().(*net.TCPAddr).Port)
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	cachePath := os.Getenv("FRAUD_POSTGRES_CACHE")
	if cachePath == "" {
		cachePath = filepath.Join(dir, "cache")
	}
	config := embeddedpostgres.DefaultConfig().Version(embeddedpostgres.V16).Port(port).
		Database("fraud_test").Username("synthetic").Password("test-password-not-a-secret").
		RuntimePath(filepath.Join(dir, "runtime")).DataPath(filepath.Join(dir, "data")).
		CachePath(cachePath).Logger(io.Discard).
		BinaryRepositoryURL("https://repo.maven.apache.org/maven2").StartTimeout(30 * time.Second)
	pg := embeddedpostgres.NewDatabase(config)
	if err := pg.Start(); err != nil {
		t.Fatalf("embedded PostgreSQL unavailable (not passed parity): %v", err)
	}
	t.Cleanup(func() {
		if err := pg.Stop(); err != nil {
			t.Errorf("stop PostgreSQL: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, config.GetConnectionURL())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { pool.Close() }()
	rules := domain.DefaultRuleConfig()
	rules.VelocityLimit = 1
	sqlStore, err := repository.NewPostgresEvents(ctx, pool, rules)
	if err != nil {
		t.Fatal(err)
	}
	var expected domain.RiskDecision
	for name, store := range map[string]repository.EventStore{"memory": repository.NewMemoryEvents(rules), "postgres": sqlStore} {
		t.Run(name, func(t *testing.T) {
			now := func() time.Time { return time.Unix(100, 0) }
			h, _ := query.NewScoreEventHandlerWithStore(store, now, time.Minute)
			h2, _ := query.NewScoreEventHandlerWithStore(store, now, time.Minute)
			input := query.ScoreEventQuery{EventID: "e", Account: "a", AmountCents: 600000}
			first, err := h.Handle(ctx, input)
			if err != nil {
				t.Fatal(err)
			}
			expected = first
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					again, err := h2.Handle(ctx, input)
					if err != nil || !reflect.DeepEqual(again, first) {
						t.Errorf("replay: %+v %v", again, err)
					}
				}()
			}
			wg.Wait()
			changed := input
			changed.AmountCents++
			if _, err := h.Handle(ctx, changed); err == nil {
				t.Fatal("conflicting replay accepted")
			}
			next, err := h.Handle(ctx, query.ScoreEventQuery{EventID: "next", Account: "a"})
			if err != nil || next.Score != rules.VelocityWeight {
				t.Fatalf("velocity %+v %v", next, err)
			}
			if err := h.FeedbackContext(ctx, "e", "CONFIRMED_FRAUD"); err != nil {
				t.Fatal(err)
			}
			if err := h2.FeedbackContext(ctx, "e", "CONFIRMED_FRAUD"); err != nil {
				t.Fatal(err)
			}
			if h.Rules().ReviewAt != 39 {
				t.Fatal("feedback applied twice")
			}
			if err := h.FeedbackContext(ctx, "e", "FALSE_POSITIVE"); err == nil {
				t.Fatal("conflicting feedback accepted")
			}
			sentinel := errors.New("rollback")
			if err := store.Update(ctx, func(s *repository.EventState) error { s.Rules.ReviewAt = 1; delete(s.Events, "e"); return sentinel }); !errors.Is(err, sentinel) {
				t.Fatal(err)
			}
			if h.Rules().ReviewAt != 39 {
				t.Fatal("transaction did not roll back")
			}
			explanation, ok, err := h.ExplainContext(ctx, "e")
			if err != nil || !ok || !reflect.DeepEqual(explanation, first) {
				t.Fatalf("explanation after rollback %+v %v", explanation, err)
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			if _, err := h.Handle(canceled, input); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation %v", err)
			}
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					_, err := h.Handle(ctx, query.ScoreEventQuery{EventID: fmt.Sprintf("parallel-%d", i), Account: "b"})
					if err != nil {
						t.Error(err)
					}
				}(i)
			}
			wg.Wait()
			if err := store.Update(ctx, func(s *repository.EventState) error {
				if len(s.Events) != 10 {
					t.Errorf("lost/duplicate events: %d", len(s.Events))
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				d, err := h.Handle(ctx, query.ScoreEventQuery{EventID: fmt.Sprintf("pattern-%d", i), Account: "structuring", AmountCents: 95000})
				if err != nil {
					t.Fatal(err)
				}
				want := 0
				if i > 0 {
					want = rules.VelocityWeight
				}
				if i == 2 {
					want += rules.StructuringWeight
				}
				if d.Score != want {
					t.Fatalf("inferred pattern parity %d: %+v", i, d)
				}
			}
		})
	}
	// Restart the actual database, not just a handler, retaining its data directory.
	pool.Close()
	if err := pg.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := pg.Start(); err != nil {
		t.Fatal(err)
	}
	pool, err = pgxpool.New(ctx, config.GetConnectionURL())
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := repository.NewPostgresEvents(ctx, pool, domain.DefaultRuleConfig())
	if err != nil {
		t.Fatal(err)
	}
	h, _ := query.NewScoreEventHandlerWithStore(reopened, time.Now, time.Minute)
	d, ok, err := h.ExplainContext(ctx, "e")
	if err != nil || !ok || !reflect.DeepEqual(d, expected) {
		t.Fatalf("database restart lost decision: %+v %v", d, err)
	}
	if h.Rules().ReviewAt != 39 {
		t.Fatal("database restart lost learned rules")
	}
	if err := h.FeedbackContext(ctx, "e", "CONFIRMED_FRAUD"); err != nil {
		t.Fatal(err)
	}
	if h.Rules().ReviewAt != 39 {
		t.Fatal("database restart lost feedback identity")
	}
	runtimeConfig, err := service.LoadConfig(func(key string) string {
		if key == "DATABASE_URL" {
			return config.GetConnectionURL()
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	selected, closeSelected, err := service.OpenStore(ctx, runtimeConfig)
	if err != nil {
		t.Fatal(err)
	}
	runtimeHandler, _ := query.NewScoreEventHandlerWithStore(selected, time.Now, time.Minute)
	if runtimeHandler.Rules().ReviewAt != 39 {
		t.Fatal("DATABASE_URL did not select durable state")
	}
	closeSelected()
	if _, err := repository.NewPostgresEvents(ctx, nil, rules); err == nil {
		t.Fatal("nil pool accepted")
	}
	invalid := rules
	invalid.ReviewAt = 0
	if _, err := repository.NewPostgresEvents(ctx, pool, invalid); err == nil {
		t.Fatal("invalid rules accepted")
	}
	if err := reopened.Update(ctx, func(s *repository.EventState) error { s.Rules.ReviewAt = 0; return nil }); err == nil {
		t.Fatal("invalid policy committed")
	}
	if h.Rules().ReviewAt != 39 {
		t.Fatal("invalid update was not rolled back")
	}
	if _, err := pool.Exec(ctx, `UPDATE fraud_state SET state='{}'::jsonb`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.LoadRules(ctx); err == nil {
		t.Fatal("corrupt stored state accepted")
	}
	pool.Close()
	if err := reopened.Update(ctx, func(*repository.EventState) error { return nil }); err == nil {
		t.Fatal("closed pool accepted transaction")
	}
	t.Log("PostgreSQL 16.9: memory/SQL replay, concurrent unique writes, rollback, cancellation, feedback, database restart parity passed")
}
