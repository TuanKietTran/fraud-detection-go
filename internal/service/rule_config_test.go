package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

const seededPolicyJSON = `{"ReviewAt":20,"HighAmountCents":100,"VeryHighAmountCents":1000,"HighAmountWeight":20}`

func policyConfig(t *testing.T, raw string) Config {
	t.Helper()
	c, err := LoadConfig(func(key string) string {
		if key == "FRAUD_RULES_JSON" {
			return raw
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestStartupRulesDefaultsOverlayAndExactIntegerBoundaries(t *testing.T) {
	defaults := domain.DefaultRuleConfig()
	for _, raw := range []string{"", " \n\t ", "{}", "{}" + strings.Repeat(" ", MaxStartupRulesBytes-2)} {
		if got := policyConfig(t, raw).InitialRules; got != defaults {
			t.Fatalf("empty policy changed defaults: %+v", got)
		}
	}
	expected := defaults
	expected.ReviewAt = 20
	expected.HighAmountCents = 100
	expected.VeryHighAmountCents = 1000
	expected.HighAmountWeight = 20
	if got := policyConfig(t, seededPolicyJSON).InitialRules; got != expected {
		t.Fatalf("partial overlay: %+v", got)
	}
	if got := policyConfig(t, `{"HighAmountWeight":0}`).InitialRules; got.HighAmountWeight != 0 {
		t.Fatal("explicit zero ignored")
	}
	// Registry must accept every canonical exported field, not just a demo subset.
	full, err := json.Marshal(defaults)
	if err != nil {
		t.Fatal(err)
	}
	if got := policyConfig(t, string(full)).InitialRules; got != defaults {
		t.Fatal("full canonical policy changed")
	}
	extreme := policyConfig(t, `{"HighAmountCents":9223372036854775807,"VeryHighAmountCents":9223372036854775807,"StructuringBandCents":9223372036854775807}`).InitialRules
	if extreme.HighAmountCents != math.MaxInt64 || extreme.VeryHighAmountCents != math.MaxInt64 || extreme.StructuringBandCents != math.MaxInt64 {
		t.Fatal("startup monetary integer lost precision")
	}
}

func TestStartupRulesRejectUnknownDuplicateNullNonIntegerOverflowAndInvalidPolicy(t *testing.T) {
	for _, raw := range []string{
		strings.Repeat(" ", MaxStartupRulesBytes+1), "null", "[]", `"object"`, "true", "42", "{", `{"ReviewAt":`,
		`{"ReviewAt":20`, `{"ReviewAt" 20}`, `{"ReviewAt":20,}`, `{} {}`, `{} null`, `{} garbage`,
		`{"reviewat":20}`, `{"review_at":20}`, `{"reviewAt":20}`, `{"Unknown":20}`, `{"ReviewAt":20,"Unknown":1}`,
		`{"ReviewAt":20,"ReviewAt":21}`, `{"ReviewAt":20,"\u0052eviewAt":21}`,
		`{"ReviewAt":null}`, `{"HighAmountCents":null}`, `{"ReviewAt":"20"}`, `{"ReviewAt":20.0}`, `{"ReviewAt":2e1}`,
		`{"ReviewAt":true}`, `{"ReviewAt":{}}`, `{"ReviewAt":[]}`, `{"ReviewAt":9223372036854775808}`, `{"HighAmountCents":9223372036854775808}`,
		`{"ReviewAt":0}`, `{"ReviewAt":71}`, `{"HighAmountWeight":101}`, `{"VelocityWeight":-1}`, `{"VelocityLimit":0}`, `{"StructuringBandCents":0}`,
		`{"HighAmountCents":-1}`, `{"VeryHighAmountCents":1}`,
	} {
		cfg, err := LoadConfig(func(key string) string {
			if key == "FRAUD_RULES_JSON" {
				return raw
			}
			return ""
		})
		if !errors.Is(err, ErrStartupRules) || cfg.InitialRules != (domain.RuleConfig{}) {
			t.Fatalf("accepted/partially exposed invalid policy %q: %+v %v", raw, cfg.InitialRules, err)
		}
	}
	marker := "synthetic-policy-input-not-a-secret"
	_, err := LoadConfig(func(key string) string {
		if key == "FRAUD_RULES_JSON" {
			return `{"` + marker + `":"` + marker + `"}`
		}
		return ""
	})
	if !errors.Is(err, ErrStartupRules) || strings.Contains(err.Error(), marker) {
		t.Fatal("startup error exposed raw policy input")
	}
	// Even callers bypassing LoadConfig cannot initialize a bad policy or listen.
	bad := policyConfig(t, "")
	bad.InitialRules.ReviewAt = 0
	bad.DatabaseURL = "postgres://127.0.0.1:1/unreachable"
	store, closeStore, err := OpenStore(context.Background(), bad)
	if !errors.Is(err, ErrStartupRules) || store != nil || closeStore != nil {
		t.Fatal("bad programmatic policy reached store initialization")
	}
	if err := Run(context.Background(), bad, quiet()); !errors.Is(err, ErrStartupRules) {
		t.Fatalf("bad policy reached serving: %v", err)
	}
}

func TestStartupPolicyDrivesMemoryHTTPScoringAndFreshRestart(t *testing.T) {
	cfg := policyConfig(t, seededPolicyJSON)
	store, closeStore, err := OpenStore(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer closeStore()
	server := NewHTTPServer(cfg, store, quiet())
	call := func(method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		server.Handler.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		if rec.Code != want {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body.String())
		}
		return rec
	}
	payload := `{"event_id":"seeded-first","account":"seeded-first","amount":"2.00","currency":"USD"}`
	first := call("POST", "/v1/score", payload, 200).Body.String()
	var result struct {
		Score    int    `json:"score"`
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal([]byte(first), &result); err != nil || result.Score != 20 || result.Decision != "REVIEW" {
		t.Fatalf("configured policy not used: %s %v", first, err)
	}
	feedback := `{"event_id":"seeded-first","label":"FALSE_POSITIVE"}`
	call("POST", "/v1/feedback", feedback, 202)
	call("POST", "/v1/feedback", feedback, 202)
	var learned domain.RuleConfig
	if err := json.Unmarshal(call("GET", "/v1/rules", "", 200).Body.Bytes(), &learned); err != nil || learned.ReviewAt != 21 {
		t.Fatal("feedback not retained exactly once", learned, err)
	}
	if got := call("POST", "/v1/score", payload, 200).Body.String(); got != first {
		t.Fatal("learned policy changed original replay")
	}
	call("POST", "/v1/feedback", `{"event_id":"seeded-first","label":"CONFIRMED_FRAUD"}`, 400)
	call("POST", "/v1/score", `{"event_id":"seeded-first","account":"seeded-first","amount":"3.00","currency":"USD"}`, 400)
	newer := call("POST", "/v1/score", `{"event_id":"seeded-new","account":"seeded-new","amount":"2.00","currency":"USD"}`, 200)
	if err := json.Unmarshal(newer.Body.Bytes(), &result); err != nil || result.Score != 20 || result.Decision != "ALLOW" {
		t.Fatal("learned threshold not applied to new scoring", result, err)
	}
	// A new memory instance is explicitly non-durable and receives its new seed.
	cfg2 := policyConfig(t, `{"ReviewAt":1,"DenyAt":1,"HighAmountCents":0,"HighAmountWeight":100}`)
	fresh, closeFresh, err := OpenStore(context.Background(), cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer closeFresh()
	server2 := NewHTTPServer(cfg2, fresh, quiet())
	rec := httptest.NewRecorder()
	server2.Handler.ServeHTTP(rec, httptest.NewRequest("GET", "/v1/events/seeded-first/explanation", nil))
	if rec.Code != 404 {
		t.Fatal("memory restart unexpectedly retained events")
	}
	rec = httptest.NewRecorder()
	server2.Handler.ServeHTTP(rec, httptest.NewRequest("POST", "/v1/score", strings.NewReader(payload)))
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || rec.Code != 200 || result.Score != 100 || result.Decision != "BLOCK" {
		t.Fatal("fresh memory seed not used", result, err)
	}
	for _, label := range []string{"FALSE_POSITIVE", "CONFIRMED_FRAUD"} {
		// Keep feedback identity immutable; use separate events for each boundary label.
		id := "boundary-" + label
		request := `{"event_id":"` + id + `","account":"` + id + `","amount":"2.00","currency":"USD"}`
		rec = httptest.NewRecorder()
		server2.Handler.ServeHTTP(rec, httptest.NewRequest("POST", "/v1/score", strings.NewReader(request)))
		if rec.Code != 200 {
			t.Fatal(rec.Body.String())
		}
		rec = httptest.NewRecorder()
		server2.Handler.ServeHTTP(rec, httptest.NewRequest("POST", "/v1/feedback", strings.NewReader(`{"event_id":"`+id+`","label":"`+label+`"}`)))
		if rec.Code != 202 {
			t.Fatal(rec.Body.String())
		}
	}
	rec = httptest.NewRecorder()
	server2.Handler.ServeHTTP(rec, httptest.NewRequest("GET", "/v1/rules", nil))
	if err := json.Unmarshal(rec.Body.Bytes(), &learned); err != nil || learned.ReviewAt != 1 || learned.Validate() != nil {
		t.Fatal("startup policy feedback broke bounds", learned, err)
	}
}
