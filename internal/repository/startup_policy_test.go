package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sqlSeedA = `{"ReviewAt":20,"HighAmountCents":100,"VeryHighAmountCents":1000,"HighAmountWeight":20}`
const sqlSeedB = `{"ReviewAt":1,"DenyAt":1,"HighAmountCents":0,"HighAmountWeight":100}`
const sqlSeedEvent = `{"event_id":"startup-original","account":"startup-original","amount":"2.00","currency":"USD"}`

func sqlPolicyConfig(t *testing.T, dsn, raw string) service.Config {
	t.Helper()
	cfg, err := service.LoadConfig(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return dsn
		case "FRAUD_RULES_JSON":
			return raw
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}
func sqlPolicyHTTP(t *testing.T, ctx context.Context, cfg service.Config) func(string, string, string, int) *httptest.ResponseRecorder {
	t.Helper()
	store, closeStore, err := service.OpenStore(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closeStore)
	server := service.NewHTTPServer(cfg, store, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	return func(method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		server.Handler.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		if rec.Code != want {
			t.Fatalf("configured SQL %s: %d %s", path, rec.Code, rec.Body.String())
		}
		return rec
	}
}
func sqlActivePolicy(t *testing.T, rec *httptest.ResponseRecorder) domain.RuleConfig {
	t.Helper()
	var rules domain.RuleConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &rules); err != nil {
		t.Fatal(err)
	}
	if err := rules.Validate(); err != nil {
		t.Fatal(err)
	}
	return rules
}
func assertSQLSeedScore(t *testing.T, rec *httptest.ResponseRecorder, score int, decision string) {
	t.Helper()
	var body struct {
		Score    int    `json:"score"`
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Score != score || body.Decision != decision {
		t.Fatalf("startup SQL scoring: %s %v", rec.Body.String(), err)
	}
}

// This uses a separate schema in the existing real-DB fixture, not another
// downloaded engine or replacement/deletion of the baseline parity state.
func setupSQLStartupPolicyBeforeRestart(t *testing.T, ctx context.Context, pool *pgxpool.Pool, dsn string) string {
	t.Helper()
	if _, err := pool.Exec(ctx, `CREATE SCHEMA startup_rules_test`); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	params := u.Query()
	params.Set("search_path", "startup_rules_test")
	u.RawQuery = params.Encode()
	dsn = u.String()
	t.Run("configured-sql-initialization-and-reopen", func(t *testing.T) {
		cfg := sqlPolicyConfig(t, dsn, sqlSeedA)
		call := sqlPolicyHTTP(t, ctx, cfg)
		if got := sqlActivePolicy(t, call("GET", "/v1/rules", "", 200)); got != cfg.InitialRules {
			t.Fatal("new SQL store did not initialize from startup policy")
		}
		original := call("POST", "/v1/score", sqlSeedEvent, 200)
		assertSQLSeedScore(t, original, 20, "REVIEW")
		feedback := `{"event_id":"startup-original","label":"FALSE_POSITIVE"}`
		call("POST", "/v1/feedback", feedback, 202)
		call("POST", "/v1/feedback", feedback, 202)
		expected := cfg.InitialRules
		expected.ReviewAt = 21
		if got := sqlActivePolicy(t, call("GET", "/v1/rules", "", 200)); got != expected {
			t.Fatal("SQL feedback not applied exactly once")
		}
		if got := call("POST", "/v1/score", sqlSeedEvent, 200).Body.String(); got != original.Body.String() {
			t.Fatal("feedback recomputed original SQL decision")
		}
		// Opening with a valid but radically different seed is not a policy update.
		reopened := sqlPolicyHTTP(t, ctx, sqlPolicyConfig(t, dsn, sqlSeedB))
		if got := sqlActivePolicy(t, reopened("GET", "/v1/rules", "", 200)); got != expected {
			t.Fatal("different startup seed reset learned SQL policy")
		}
		assertSQLSeedScore(t, reopened("POST", "/v1/score", `{"event_id":"new-before-restart","account":"new-before-restart","amount":"2.00","currency":"USD"}`, 200), 20, "ALLOW")
	})
	return dsn
}

func assertSQLStartupPolicyAfterRestart(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	t.Run("configured-sql-database-restart-retains-learned-policy", func(t *testing.T) {
		expected := sqlPolicyConfig(t, dsn, sqlSeedA).InitialRules
		expected.ReviewAt = 21
		for _, raw := range []string{sqlSeedB, ""} {
			call := sqlPolicyHTTP(t, ctx, sqlPolicyConfig(t, dsn, raw))
			if got := sqlActivePolicy(t, call("GET", "/v1/rules", "", 200)); got != expected {
				t.Fatal("database restart reapplied a startup seed/default")
			}
			assertSQLSeedScore(t, call("POST", "/v1/score", sqlSeedEvent, 200), 20, "REVIEW")
			call("POST", "/v1/feedback", `{"event_id":"startup-original","label":"FALSE_POSITIVE"}`, 202)
			call("POST", "/v1/feedback", `{"event_id":"startup-original","label":"CONFIRMED_FRAUD"}`, 400)
			call("POST", "/v1/score", `{"event_id":"startup-original","account":"startup-original","amount":"3.00","currency":"USD"}`, 400)
			if got := sqlActivePolicy(t, call("GET", "/v1/rules", "", 200)); got != expected {
				t.Fatal("restart lost feedback identity/validation")
			}
		}
		call := sqlPolicyHTTP(t, ctx, sqlPolicyConfig(t, dsn, sqlSeedB))
		assertSQLSeedScore(t, call("POST", "/v1/score", `{"event_id":"new-after-restart","account":"new-after-restart","amount":"2.00","currency":"USD"}`, 200), 20, "ALLOW")
		invalid := sqlPolicyConfig(t, dsn, sqlSeedB)
		invalid.InitialRules.ReviewAt = 0
		store, closeStore, err := service.OpenStore(ctx, invalid)
		if !errors.Is(err, service.ErrStartupRules) || store != nil || closeStore != nil {
			t.Fatal("existing SQL state hid invalid incoming startup policy")
		}
		if got := sqlActivePolicy(t, call("GET", "/v1/rules", "", 200)); got != expected {
			t.Fatal("invalid seed changed durable state")
		}
	})
}
