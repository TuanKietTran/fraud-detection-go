package api

import (
	"encoding/json"
	"net/http"

	"github.com/example/fraud-detection/internal/query"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	score *query.ScoreEventHandler
}

func NewHandlers(s *query.ScoreEventHandler) *Handlers {
	return &Handlers{score: s}
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/score", h.scoreEvent)
	return r
}

type scoreReq struct {
	EventID string  `json:"event_id"`
	Amount  float64 `json:"amount"`
	Account string  `json:"account"`
}

func (h *Handlers) scoreEvent(w http.ResponseWriter, r *http.Request) {
	var req scoreReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	res, err := h.score.Handle(r.Context(), query.ScoreEventQuery{
		EventID: req.EventID,
		Amount:  req.Amount,
		Account: req.Account,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
