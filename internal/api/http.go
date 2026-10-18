package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/TuanKietTran/fraud-detection-go/internal/query"
)

const maxBodyBytes = 1 << 20

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	RequestID string `json:"request_id"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	body := errorBody{RequestID: w.Header().Get("X-Request-ID")}
	body.Error.Code = code
	body.Error.Message = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func useCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, query.ErrInvalid):
		writeError(w, 400, "invalid_request", err.Error())
	case errors.Is(err, query.ErrConflict):
		writeError(w, 400, "payload_conflict", err.Error()) // documented compatibility status
	case errors.Is(err, query.ErrNotFound):
		writeError(w, 404, "not_found", "event not found")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		writeError(w, 408, "request_timeout", "request canceled or timed out")
	default:
		writeError(w, 503, "unavailable", "scoring store unavailable")
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(out); err != nil {
		bodyError(w, err)
		return false
	}
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return true
	}
	if err != nil {
		bodyError(w, err)
	} else {
		writeError(w, 400, "invalid_json", "exactly one JSON value required")
	}
	return false
}

func bodyError(w http.ResponseWriter, err error) {
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		writeError(w, 413, "body_too_large", "request body exceeds 1048576 bytes")
		return
	}
	writeError(w, 400, "invalid_json", "invalid json")
}

func (h *Handlers) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate an authoritative bounded ID; never reflect attacker-controlled headers.
		id := h.nextID()
		w.Header().Set("X-Request-ID", id)
		defer func() {
			if recover() != nil {
				h.logger.Error("http panic recovered", slog.String("request_id", id))
				writeError(w, 500, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
