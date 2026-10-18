package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
	"github.com/TuanKietTran/fraud-detection-go/internal/query"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	score  *query.ScoreEventHandler
	nextID func() string
	logger *slog.Logger
}

func NewHandlers(s *query.ScoreEventHandler) *Handlers {
	return &Handlers{score: s, nextID: uuid.NewString, logger: slog.Default()}
}
func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.protect)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeError(w, 404, "not_found", "route not found") })
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, 405, "method_not_allowed", "method not allowed")
	})
	r.Post("/score", h.scoreEvent)
	r.Get("/events/{event_id}/explanation", h.explanation)
	r.Get("/rules", h.rules)
	r.Post("/feedback", h.feedback)
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
type feedbackRequest struct {
	EventID string `json:"event_id"`
	Label   string `json:"label"`
}
type scoreResponse struct {
	EventID  string          `json:"event_id"`
	Score    int             `json:"score"`
	Decision domain.Decision `json:"decision"`
	Reasons  []string        `json:"reasons"`
}

func (h *Handlers) explanation(w http.ResponseWriter, r *http.Request) {
	result, ok, err := h.score.ExplainContext(r.Context(), chi.URLParam(r, "event_id"))
	if err != nil {
		useCaseError(w, err)
		return
	}
	if !ok {
		writeError(w, 404, "not_found", "event not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(scoreResponse{result.EventID, result.Score, result.Decision, result.ReasonCodes})
}
func (h *Handlers) rules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.score.LoadRules(r.Context())
	if err != nil {
		useCaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rules)
}
func (h *Handlers) feedback(w http.ResponseWriter, r *http.Request) {
	var req feedbackRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if err := h.score.FeedbackContext(r.Context(), req.EventID, req.Label); err != nil {
		// Preserve legacy feedback 400 for absent events as published.
		if errors.Is(err, query.ErrNotFound) {
			writeError(w, 400, "not_found", "event not found")
			return
		}
		useCaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
func (h *Handlers) scoreEvent(w http.ResponseWriter, r *http.Request) {
	var req scoreReq
	if !decodeBody(w, r, &req) {
		return
	}
	if req.EventID == "" || req.Account == "" {
		writeError(w, 400, "invalid_request", "event_id and account required")
		return
	}
	amountText := strings.TrimSpace(string(req.Amount))
	if len(amountText) > 1 && amountText[0] == '"' {
		if err := json.Unmarshal(req.Amount, &amountText); err != nil {
			writeError(w, 400, "invalid_amount", "invalid amount")
			return
		}
	}
	if req.Currency == "" {
		req.Currency = "USD"
	} // legacy compatibility
	cents, err := domain.ParseMinorUnits(amountText, req.Currency)
	if err != nil {
		writeError(w, 400, "invalid_amount", err.Error())
		return
	}
	res, err := h.score.Handle(r.Context(), query.ScoreEventQuery{EventID: req.EventID, AmountCents: cents, Account: req.Account, RecentCount: req.RecentCount, GeoMismatch: req.GeoMismatch, DeviceMismatch: req.DeviceMismatch, NewPayee: req.NewPayee, Structuring: req.Structuring})
	if err != nil {
		useCaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	decision := res.Decision
	if decision == domain.Deny {
		decision = domain.Block
	}
	_ = json.NewEncoder(w).Encode(scoreResponse{res.EventID, res.Score, decision, res.ReasonCodes})
}
