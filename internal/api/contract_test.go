package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TuanKietTran/fraud-detection-go/internal/query"
)

func TestPublishedImplementedRouteAndStatuses(t *testing.T) {
	h := NewHandlers(query.NewScoreEventHandler())
	for _, tc := range []struct {
		body string
		want int
	}{
		{`{"event_id":"evt-1","account":"acct-1","amount":100}`, http.StatusOK},
		{`{`, http.StatusBadRequest},
	} {
		req := httptest.NewRequest(http.MethodPost, "/score", strings.NewReader(tc.body))
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("status=%d want %d body=%s", rec.Code, tc.want, rec.Body.String())
		}
	}
}
