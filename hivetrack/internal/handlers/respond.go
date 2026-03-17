package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/the127/hivetrack/internal/models"
)

type errorResponse struct {
	Errors []apiError `json:"errors"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// RespondJSON writes a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// RespondError maps known errors to HTTP status codes.
func RespondError(w http.ResponseWriter, err error) {
	var status int
	var code string

	switch {
	case errors.Is(err, models.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, models.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
	case errors.Is(err, models.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
	case errors.Is(err, models.ErrBadRequest):
		status = http.StatusBadRequest
		code = "bad_request"
	default:
		status = http.StatusInternalServerError
		code = "internal"
	}

	RespondJSON(w, status, errorResponse{
		Errors: []apiError{{Code: code, Message: err.Error()}},
	})
}
