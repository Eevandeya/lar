package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eevandeya/lar/internal/api"
)

func writeError(w http.ResponseWriter, statusCode int, code api.ErrorCode, message string) error {
	respErr := api.ErrorResponse{
		Error: &api.Error{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", api.ApplicationJSON)
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(respErr)
}

func writeAuthError(w http.ResponseWriter) {
	if err := writeError(w, http.StatusUnauthorized, api.ErrUnauthorized, "invalid credentials"); err != nil {
		slog.Error("failed to write error response", "err", err)
	}
}
