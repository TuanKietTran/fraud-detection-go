package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/go-chi/chi/v5"
)

type Handlers struct{ score *query.ScoreEventHandler }

func NewHandlers(s *query.ScoreEventHandler) *Handlers { return &Handlers{score: s} }
func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/score", h.scoreEvent)
	return r
}

type scoreReq struct {
	EventID        string          `json:"event_id"`
	Amount         json.RawMessage `json:"amount"`
	Currency       string          `json:"currency"`
	Account        string          `json:"account"`
	RecentCount    int             `json:"recent_count"`
	GeoMismatch    bool            `json:"geo_mismatch"`
	DeviceMismatch bool            `json:"device_mismatch"`
	NewPayee       bool            `json:"new_payee"`
	Structuring    bool            `json:"structuring"`
}
type scoreResponse struct {
	EventID  string          `json:"event_id"`
	Score    int             `json:"score"`
	Decision domain.Decision `json:"decision"`
	Reasons  []string        `json:"reasons"`
}

func (h *Handlers) scoreEvent(w http.ResponseWriter, r *http.Request) {
	var req scoreReq
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		http.Error(w, "invalid json", 400)
		return
	}
	amountText := strings.TrimSpace(string(req.Amount))
	if len(amountText) > 1 && amountText[0] == '"' {
		if err := json.Unmarshal(req.Amount, &amountText); err != nil {
			http.Error(w, "invalid amount", 400)
			return
		}
	}
	if req.Currency == "" {
		req.Currency = "USD"
	} // legacy compatibility
	cents, err := domain.ParseMinorUnits(amountText, req.Currency)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	res, err := h.score.Handle(r.Context(), query.ScoreEventQuery{EventID: req.EventID, AmountCents: cents, Account: req.Account, RecentCount: req.RecentCount, GeoMismatch: req.GeoMismatch, DeviceMismatch: req.DeviceMismatch, NewPayee: req.NewPayee, Structuring: req.Structuring})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	decision := res.Decision
	if decision == domain.Deny {
		decision = domain.Block
	}
	_ = json.NewEncoder(w).Encode(scoreResponse{res.EventID, res.Score, decision, res.ReasonCodes})
}
