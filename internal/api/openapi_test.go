package api

import (
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestOpenAPIParsedRoutesAndStatuses(t *testing.T) {
	b, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Summary   string         `yaml:"summary"`
			Responses map[string]any `yaml:"responses"`
		} `yaml:"paths"`
	}
	if err = yaml.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	op, ok := doc.Paths["/v1/score"]["post"]
	if !ok {
		t.Fatal("missing POST /v1/score")
	}
	if len(op.Responses) != 6 {
		t.Errorf("score status set drift: %v", op.Responses)
	}
	for _, status := range []string{"200", "400", "408", "413", "500", "503"} {
		if _, ok := op.Responses[status]; !ok {
			t.Errorf("OpenAPI missing status %s", status)
		}
	}
	for _, route := range []struct {
		path, method string
		statuses     []string
	}{{"/v1/events/{event_id}/explanation", "get", []string{"200", "404", "408", "500", "503"}}, {"/v1/rules", "get", []string{"200", "408", "500", "503"}}, {"/v1/feedback", "post", []string{"202", "400", "408", "413", "500", "503"}}} {
		operation, ok := doc.Paths[route.path][route.method]
		if !ok {
			t.Errorf("OpenAPI missing %s %s", route.method, route.path)
			continue
		}
		if len(operation.Responses) != len(route.statuses) {
			t.Errorf("%s status set drift: %v", route.path, operation.Responses)
		}
		for _, status := range route.statuses {
			if _, ok := operation.Responses[status]; !ok {
				t.Errorf("OpenAPI missing %s %s", route.path, status)
			}
		}
	}
	h := NewHandlers(query.NewScoreEventHandler())
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{{http.MethodPost, "/score", `{"event_id":"e","account":"a","amount":"6000.00","currency":"USD"}`, 200}, {http.MethodPost, "/score", `{"event_id":"e","account":"a","amount":"6000.00","currency":"USD"}`, 200}, {http.MethodPost, "/score", `{"event_id":"e","account":"a","amount":"6000.01","currency":"USD"}`, 400}, {http.MethodPost, "/score", `{"event_id":"bad","account":"a","amount":"1.001","currency":"USD"}`, 400}, {http.MethodGet, "/rules", "", 200}, {http.MethodGet, "/events/e/explanation", "", 200}, {http.MethodPost, "/feedback", `{"event_id":"e","label":"CONFIRMED_FRAUD"}`, 202}, {http.MethodGet, "/events/missing/explanation", "", 404}} {
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if rec.Code != tc.want {
			t.Errorf("actual POST /v1/score status=%d want=%d", rec.Code, tc.want)
		}
	}
}
