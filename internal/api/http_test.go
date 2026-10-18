package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/TuanKietTran/fraud-detection-go/internal/repository"
	"time"
)

type brokenStore struct {
	err        error
	panicValue bool
}

func (s brokenStore) Update(context.Context, func(*repository.EventState) error) error {
	if s.panicValue {
		panic("synthetic sensitive panic must not be exposed")
	}
	return s.err
}

func TestJSONErrorsBodyLimitsRequestIDsAndRecovery(t *testing.T) {
	valid := `{"event_id":"e","account":"a","amount":"1","currency":"USD"}`
	for _, tc := range []struct {
		name, method, path, body, code string
		status                         int
	}{
		{"malformed", "POST", "/score", "{", "invalid_json", 400},
		{"trailing-score", "POST", "/score", valid + ` {}`, "invalid_json", 400},
		{"trailing-feedback", "POST", "/feedback", `{"event_id":"e","label":"CONFIRMED_FRAUD"} {}`, "invalid_json", 400},
		{"oversized-score", "POST", "/score", `{"event_id":"` + strings.Repeat("x", maxBodyBytes) + `"}`, "body_too_large", 413},
		{"oversized-feedback", "POST", "/feedback", `{"event_id":"` + strings.Repeat("x", maxBodyBytes) + `"}`, "body_too_large", 413},
		{"invalid-money", "POST", "/score", `{"event_id":"e","account":"a","amount":"1.001","currency":"USD"}`, "invalid_amount", 400},
		{"missing-fields", "POST", "/score", `{}`, "invalid_request", 400},
		{"absent", "GET", "/events/missing/explanation", "", "not_found", 404},
		{"unknown", "GET", "/missing", "", "not_found", 404},
		{"method", "DELETE", "/score", "", "method_not_allowed", 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandlers(query.NewScoreEventHandler())
			h.nextID = func() string { return "test-request-id" }
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("X-Request-ID", "untrusted")
			rec := httptest.NewRecorder()
			h.Routes().ServeHTTP(rec, req)
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if rec.Code != tc.status || body.Error.Code != tc.code || body.RequestID != "test-request-id" || rec.Header().Get("X-Request-ID") != body.RequestID || rec.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("%d %+v", rec.Code, body)
			}
		})
	}
}

func TestStoreFailuresAndCancellationNeverBecomeSuccessfulResponses(t *testing.T) {
	for _, tc := range []struct {
		store brokenStore
		want  int
		code  string
	}{
		{brokenStore{err: errors.New("sensitive database internals")}, 503, "unavailable"},
		{brokenStore{err: context.Canceled}, 408, "request_timeout"},
		{brokenStore{err: context.DeadlineExceeded}, 408, "request_timeout"},
		{brokenStore{panicValue: true}, 500, "internal_error"},
	} {
		for _, route := range []struct{ method, path, body string }{
			{"POST", "/score", `{"event_id":"e","account":"a","amount":"1","currency":"USD"}`},
			{"GET", "/rules", ""}, {"GET", "/events/e/explanation", ""},
			{"POST", "/feedback", `{"event_id":"e","label":"CONFIRMED_FRAUD"}`},
		} {
			handler, _ := query.NewScoreEventHandlerWithStore(tc.store, time.Now, time.Minute)
			rec := httptest.NewRecorder()
			NewHandlers(handler).Routes().ServeHTTP(rec, httptest.NewRequest(route.method, route.path, strings.NewReader(route.body)))
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if rec.Code != tc.want || body.Error.Code != tc.code || strings.Contains(rec.Body.String(), "sensitive") || body.RequestID == "" {
				t.Fatalf("%s: %d %s", route.path, rec.Code, rec.Body.String())
			}
		}
	}
}

func TestFeedbackJSONLegacyErrorsAndReplay(t *testing.T) {
	h := NewHandlers(query.NewScoreEventHandler())
	for _, tc := range []struct {
		path, body string
		want       int
		code       string
	}{
		{"/feedback", `{"event_id":"e","label":"CONFIRMED_FRAUD"}`, 400, "not_found"},
		{"/score", `{"event_id":"e","account":"a","amount":"1"}`, 200, ""},
		{"/feedback", `{"event_id":"e","label":"invalid"}`, 400, "invalid_request"},
		{"/feedback", `{"event_id":"e","label":"CONFIRMED_FRAUD"}`, 202, ""},
		{"/feedback", `{"event_id":"e","label":"CONFIRMED_FRAUD"}`, 202, ""},
		{"/feedback", `{"event_id":"e","label":"FALSE_POSITIVE"}`, 400, "payload_conflict"},
		{"/score", `{"event_id":"e","account":"a","amount":"2"}`, 400, "payload_conflict"},
	} {
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body)))
		if rec.Code != tc.want {
			t.Fatalf("%d %s", rec.Code, rec.Body.String())
		}
		if tc.code != "" {
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code != tc.code {
				t.Fatalf("%+v %v", body, err)
			}
		}
	}
}
