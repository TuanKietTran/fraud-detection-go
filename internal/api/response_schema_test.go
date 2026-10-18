package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"gopkg.in/yaml.v3"
)

// Targeted response-contract subset, not a general OpenAPI/JSON Schema engine.
// JSON numbers remain exact text and are checked as int64, never float money.
type responseSchema struct {
	Ref        string                     `yaml:"$ref"`
	Type       string                     `yaml:"type"`
	Required   []string                   `yaml:"required"`
	Properties map[string]*responseSchema `yaml:"properties"`
	Additional *bool                      `yaml:"additionalProperties"`
	Enum       []string                   `yaml:"enum"`
	MinLength  *int                       `yaml:"minLength"`
	Minimum    *int64                     `yaml:"minimum"`
	Maximum    *int64                     `yaml:"maximum"`
	Items      *responseSchema            `yaml:"items"`
}
type contractResponse struct {
	Ref     string `yaml:"$ref"`
	Content map[string]struct {
		Schema *responseSchema `yaml:"schema"`
	} `yaml:"content"`
}
type responseContract struct {
	Paths map[string]map[string]struct {
		Responses map[string]*contractResponse `yaml:"responses"`
	} `yaml:"paths"`
	Components struct {
		Schemas   map[string]*responseSchema   `yaml:"schemas"`
		Responses map[string]*contractResponse `yaml:"responses"`
	} `yaml:"components"`
}

func loadResponseContract(t *testing.T) *responseContract {
	t.Helper()
	data, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc responseContract
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return &doc
}
func (d *responseContract) schema(s *responseSchema) (*responseSchema, error) {
	for hops := 0; hops < 16; hops++ {
		if s == nil {
			return nil, fmt.Errorf("missing schema")
		}
		if s.Ref == "" {
			return s, nil
		}
		const prefix = "#/components/schemas/"
		if !strings.HasPrefix(s.Ref, prefix) {
			return nil, fmt.Errorf("unsupported schema reference %s", s.Ref)
		}
		s = d.Components.Schemas[strings.TrimPrefix(s.Ref, prefix)]
	}
	return nil, fmt.Errorf("cyclic schema alias")
}
func (d *responseContract) response(r *contractResponse) (*contractResponse, error) {
	for hops := 0; hops < 16; hops++ {
		if r == nil {
			return nil, fmt.Errorf("missing response")
		}
		if r.Ref == "" {
			return r, nil
		}
		const prefix = "#/components/responses/"
		if !strings.HasPrefix(r.Ref, prefix) {
			return nil, fmt.Errorf("unsupported response reference %s", r.Ref)
		}
		r = d.Components.Responses[strings.TrimPrefix(r.Ref, prefix)]
	}
	return nil, fmt.Errorf("cyclic response alias")
}
func (d *responseContract) validate(s *responseSchema, value any) error {
	s, err := d.schema(s)
	if err != nil {
		return err
	}
	switch s.Type {
	case "object":
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object")
		}
		for _, key := range s.Required {
			if _, ok := obj[key]; !ok {
				return fmt.Errorf("missing required %s", key)
			}
		}
		for key, v := range obj {
			property, known := s.Properties[key]
			if !known {
				if s.Additional != nil && !*s.Additional {
					return fmt.Errorf("unexpected property %s", key)
				}
				continue
			}
			if err := d.validate(property, v); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	case "array":
		values, ok := value.([]any)
		if !ok {
			return fmt.Errorf("expected array")
		}
		for _, v := range values {
			if err := d.validate(s.Items, v); err != nil {
				return err
			}
		}
	case "string":
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}
		if s.MinLength != nil && utf8.RuneCountInString(text) < *s.MinLength {
			return fmt.Errorf("string too short")
		}
		if len(s.Enum) > 0 {
			found := false
			for _, allowed := range s.Enum {
				if text == allowed {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("unknown enum %s", text)
			}
		}
	case "integer":
		number, ok := value.(json.Number)
		if !ok {
			return fmt.Errorf("expected exact JSON integer")
		}
		integer, err := number.Int64()
		if err != nil {
			return fmt.Errorf("invalid int64: %w", err)
		}
		if s.Minimum != nil && integer < *s.Minimum {
			return fmt.Errorf("below minimum")
		}
		if s.Maximum != nil && integer > *s.Maximum {
			return fmt.Errorf("above maximum")
		}
	default:
		return fmt.Errorf("unsupported response schema type %q", s.Type)
	}
	return nil
}
func decodeExactJSON(body string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("expected one JSON value")
	}
	return value, nil
}
func (d *responseContract) assertResponse(t *testing.T, method, path string, rec *httptest.ResponseRecorder) {
	t.Helper()
	operation, ok := d.Paths[path][strings.ToLower(method)]
	if !ok {
		t.Fatalf("undeclared operation %s %s", method, path)
	}
	definition, err := d.response(operation.Responses[strconv.Itoa(rec.Code)])
	if err != nil {
		t.Fatal(err)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request ID")
	}
	content, hasJSON := definition.Content["application/json"]
	if !hasJSON {
		if rec.Body.Len() != 0 {
			t.Fatalf("expected empty response, got %s", rec.Body.String())
		}
		return
	}
	media, _, err := mime.ParseMediaType(rec.Header().Get("Content-Type"))
	if err != nil || media != "application/json" {
		t.Fatalf("wrong media type: %s", rec.Header().Get("Content-Type"))
	}
	value, err := decodeExactJSON(rec.Body.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := d.validate(content.Schema, value); err != nil {
		t.Fatalf("%s %s status%d: %v; body=%s", method, path, rec.Code, err, rec.Body.String())
	}
	if rec.Code >= 400 && value.(map[string]any)["request_id"] != rec.Header().Get("X-Request-ID") {
		t.Fatal("error/header request ID mismatch")
	}
}

func TestV1ResponsesMatchParsedSchemasForEveryDeclaredStatus(t *testing.T) {
	doc := loadResponseContract(t)
	h := NewHandlers(query.NewScoreEventHandler())
	seen := make(map[string]bool)
	seenReasons := make(map[string]bool)
	call := func(handler *Handlers, method, path, contractPath, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		if rec.Code != want {
			t.Fatalf("%s %s: %d want%d: %s", method, path, rec.Code, want, rec.Body.String())
		}
		doc.assertResponse(t, method, contractPath, rec)
		seen[strings.ToLower(method)+" "+contractPath+" "+strconv.Itoa(want)] = true
		if want == 200 && (contractPath == "/v1/score" || contractPath == "/v1/events/{event_id}/explanation") {
			parsed, _ := decodeExactJSON(rec.Body.String())
			for _, reason := range parsed.(map[string]any)["reasons"].([]any) {
				seenReasons[reason.(string)] = true
			}
		}
		return rec
	}
	for _, tc := range []struct{ id, account, amount, decision string }{{"allow", "a", "1", "ALLOW"}, {"review", "r", "6000", "REVIEW"}, {"deny", "d", "20000", "BLOCK"}} {
		payload := fmt.Sprintf(`{"event_id":%q,"account":%q,"amount":%q,"currency":"USD"}`, tc.id, tc.account, tc.amount)
		score := call(h, "POST", "/score", "/v1/score", payload, 200)
		result, _ := decodeExactJSON(score.Body.String())
		if result.(map[string]any)["decision"] != tc.decision {
			t.Fatal("score decision drift")
		}
		explain := call(h, "GET", "/events/"+tc.id+"/explanation", "/v1/events/{event_id}/explanation", "", 200)
		expected := tc.decision
		if expected == "BLOCK" {
			expected = "DENY"
		}
		result, _ = decodeExactJSON(explain.Body.String())
		if result.(map[string]any)["decision"] != expected {
			t.Fatal("canonical explanation decision drift")
		}
	}
	call(h, "POST", "/score", "/v1/score", `{"event_id":"signals-high","account":"signals-high","amount":"20000","currency":"USD","recent_count":5,"geo_mismatch":true,"device_mismatch":true,"new_payee":true}`, 200)
	call(h, "POST", "/score", "/v1/score", `{"event_id":"signals-band","account":"signals-band","amount":"950","currency":"USD","structuring":true}`, 200)
	for _, reason := range doc.Components.Schemas["Reasons"].Items.Enum {
		if !seenReasons[reason] {
			t.Errorf("unexercised declared reason %s", reason)
		}
	}
	rules := call(h, "GET", "/rules", "/v1/rules", "", 200)
	initial, _ := decodeExactJSON(rules.Body.String())
	if initial.(map[string]any)["ReviewAt"] != json.Number("40") {
		t.Fatal("unexpected default threshold")
	}
	call(h, "POST", "/feedback", "/v1/feedback", `{"event_id":"review","label":"CONFIRMED_FRAUD"}`, 202)
	call(h, "POST", "/feedback", "/v1/feedback", `{"event_id":"review","label":"CONFIRMED_FRAUD"}`, 202)
	rules = call(h, "GET", "/rules", "/v1/rules", "", 200)
	learned, _ := decodeExactJSON(rules.Body.String())
	if learned.(map[string]any)["ReviewAt"] != json.Number("39") {
		t.Fatal("learned threshold drift")
	}
	call(h, "POST", "/feedback", "/v1/feedback", `{"event_id":"review","label":"FALSE_POSITIVE"}`, 400)
	call(h, "POST", "/score", "/v1/score", `{"event_id":"allow","account":"a","amount":"2"}`, 400)
	call(h, "GET", "/events/missing/explanation", "/v1/events/{event_id}/explanation", "", 404)
	for _, route := range []struct{ method, path, contractPath, payload string }{
		{"POST", "/score", "/v1/score", `{"event_id":"fault","account":"f","amount":"1"}`},
		{"GET", "/rules", "/v1/rules", ""},
		{"GET", "/events/fault/explanation", "/v1/events/{event_id}/explanation", ""},
		{"POST", "/feedback", "/v1/feedback", `{"event_id":"fault","label":"CONFIRMED_FRAUD"}`},
	} {
		for _, fault := range []struct {
			store  brokenStore
			status int
		}{{brokenStore{err: context.Canceled}, 408}, {brokenStore{err: errors.New("synthetic unavailable")}, 503}, {brokenStore{panicValue: true}, 500}} {
			scorer, err := query.NewScoreEventHandlerWithStore(fault.store, time.Now, time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			handler := NewHandlers(scorer)
			call(handler, route.method, route.path, route.contractPath, route.payload, fault.status)
		}
	}
	huge := `{"event_id":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	call(h, "POST", "/score", "/v1/score", huge, 413)
	call(h, "POST", "/feedback", "/v1/feedback", huge, 413)
	for path, methods := range doc.Paths {
		if strings.HasPrefix(path, "/v1/") {
			for method, operation := range methods {
				for status := range operation.Responses {
					if !seen[method+" "+path+" "+status] {
						t.Errorf("unexercised declared v1 response %s %s %s", method, path, status)
					}
				}
			}
		}
	}
}

func TestRuleConfigSchemaMatchesExtremeExactIntegerSerialization(t *testing.T) {
	doc := loadResponseContract(t)
	config := domain.DefaultRuleConfig()
	config.HighAmountCents = math.MaxInt64
	config.VeryHighAmountCents = math.MaxInt64
	config.StructuringBandCents = 1
	config.ReviewAt = 1
	config.DenyAt = 100
	config.HighAmountWeight = 0
	config.VelocityWeight = 100
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
	h := NewHandlers(query.NewScoreEventHandlerWithConfig(config))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest("GET", "/rules", nil))
	doc.assertResponse(t, "GET", "/v1/rules", rec)
	parsed, err := decodeExactJSON(rec.Body.String())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.(map[string]any)["HighAmountCents"] != json.Number("9223372036854775807") {
		t.Fatal("int64 monetary configuration lost precision")
	}
	// The schema must enumerate exactly every serialized policy field as required.
	schema, err := doc.schema(doc.Components.Schemas["RuleConfig"])
	if err != nil {
		t.Fatal(err)
	}
	fields := parsed.(map[string]any)
	if len(schema.Properties) != len(fields) || len(schema.Required) != len(fields) {
		t.Fatal("incomplete rule field/required schema")
	}
	for _, key := range schema.Required {
		if _, ok := schema.Properties[key]; !ok {
			t.Fatalf("required property %s has no schema", key)
		}
	}
}

func TestResponseSchemaChecksRejectPayloadDrift(t *testing.T) {
	doc := loadResponseContract(t)
	for _, tc := range []struct{ schema, body string }{
		{"ScoreResponse", `{"event_id":"e","score":101,"decision":"ALLOW","reasons":[]}`},
		{"ScoreResponse", `{"event_id":"e","score":-1,"decision":"ALLOW","reasons":[]}`},
		{"ScoreResponse", `{"event_id":"e","score":1.5,"decision":"ALLOW","reasons":[]}`},
		{"ScoreResponse", `{"event_id":"e","score":"1","decision":"ALLOW","reasons":[]}`},
		{"ScoreResponse", `{"event_id":"","score":0,"decision":"ALLOW","reasons":[]}`},
		{"ScoreResponse", `{"event_id":"e","score":0,"decision":"UNKNOWN","reasons":[]}`},
		{"ScoreResponse", `{"event_id":"e","score":0,"decision":"ALLOW","reasons":["UNKNOWN"]}`},
		{"ScoreResponse", `{"event_id":"e","score":0,"decision":"ALLOW","reasons":null}`},
		{"ScoreResponse", `{"event_id":"e","score":0,"decision":"ALLOW"}`},
		{"ScoreResponse", `{"event_id":"e","score":0,"decision":"ALLOW","reasons":[],"extra":1}`},
		{"ExplanationResponse", `{"event_id":"e","score":70,"decision":"BLOCK","reasons":[]}`},
		{"Error", `{"error":{"message":"bad"},"request_id":"id"}`},
		{"Error", `{"error":{"code":"bad","message":"bad"},"request_id":""}`},
		{"Error", `{"error":{"code":"bad","message":"bad","stack":"internal"},"request_id":"id"}`},
		{"RuleConfig", `{"review_at":40}`},
	} {
		value, err := decodeExactJSON(tc.body)
		if err != nil {
			t.Fatal(err)
		}
		if err := doc.validate(doc.Components.Schemas[tc.schema], value); err == nil {
			t.Errorf("accepted %s drift: %s", tc.schema, tc.body)
		}
	}
	seed, err := json.Marshal(domain.DefaultRuleConfig())
	if err != nil {
		t.Fatal(err)
	}
	for key, bad := range map[string]json.Number{"ReviewAt": "0", "DenyAt": "101", "HighAmountWeight": "101", "VelocityWeight": "-1", "VelocityLimit": "0", "HighAmountCents": "9223372036854775808", "StructuringBandCents": "0"} {
		value, err := decodeExactJSON(string(seed))
		if err != nil {
			t.Fatal(err)
		}
		value.(map[string]any)[key] = bad
		if err := doc.validate(doc.Components.Schemas["RuleConfig"], value); err == nil {
			t.Errorf("accepted invalid RuleConfig %s=%s", key, bad)
		}
	}
	if _, err := decodeExactJSON(`{} {}`); err == nil {
		t.Fatal("accepted multiple JSON values")
	}
}

func TestPublishedResponseReferencesAndSchemaDefinitionsResolve(t *testing.T) {
	doc := loadResponseContract(t)
	visited := make(map[*responseSchema]bool)
	var walk func(*responseSchema)
	walk = func(s *responseSchema) {
		resolved, err := doc.schema(s)
		if err != nil {
			t.Fatal(err)
		}
		if visited[resolved] {
			return
		}
		visited[resolved] = true
		if resolved.Type == "object" {
			for _, property := range resolved.Properties {
				walk(property)
			}
		}
		if resolved.Type == "array" {
			walk(resolved.Items)
		}
	}
	for _, schema := range doc.Components.Schemas {
		walk(schema)
	}
	for path, methods := range doc.Paths {
		for method, operation := range methods {
			for status, response := range operation.Responses {
				resolved, err := doc.response(response)
				if err != nil {
					t.Fatalf("%s %s %s: %v", method, path, status, err)
				}
				for _, media := range resolved.Content {
					walk(media.Schema)
				}
				number, err := strconv.Atoi(status)
				if err != nil {
					t.Fatal(err)
				}
				if number >= 400 {
					if resolved.Content["application/json"].Schema == nil {
						t.Fatalf("missing JSON error schema %s %s %s", method, path, status)
					}
				}
			}
		}
	}
	if _, err := doc.schema(&responseSchema{Ref: "#/components/schemas/does-not-exist"}); err == nil {
		t.Fatal("accepted dangling schema")
	}
	if _, err := doc.response(&contractResponse{Ref: "#/components/responses/does-not-exist"}); err == nil {
		t.Fatal("accepted dangling response")
	}
}
