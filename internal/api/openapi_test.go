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
	for _, status := range []string{"200", "400"} {
		if _, ok := op.Responses[status]; !ok {
			t.Errorf("OpenAPI missing status %s", status)
		}
	}
	h := NewHandlers(query.NewScoreEventHandler())
	for _, tc := range []struct {
		body string
		want int
	}{{`{"event_id":"e","account":"a","amount":"6000.00","currency":"USD"}`, 200}, {`{"event_id":"e","account":"a","amount":"1.001","currency":"USD"}`, 400}} {
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/score", strings.NewReader(tc.body)))
		if rec.Code != tc.want {
			t.Errorf("actual POST /v1/score status=%d want=%d", rec.Code, tc.want)
		}
	}
}
