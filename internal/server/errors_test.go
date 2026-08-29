package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/stretchr/testify/require"
)

func TestWriteError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		errorMessage := "invalid machine name"
		rec := httptest.NewRecorder()
		err := writeError(rec, http.StatusBadRequest, api.ErrInvalidMachineName, errorMessage)
		require.NoError(t, err)

		require.Equal(t, api.ApplicationJSON, rec.Header().Get("Content-Type"))
		require.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp api.ErrorResponse
		err = json.NewDecoder(rec.Body).Decode(&errorResp)
		require.NoError(t, err)

		require.Equal(t, api.ErrInvalidMachineName, errorResp.Error.Code)
		require.Equal(t, errorMessage, errorResp.Error.Message)
	})
}

func TestWriteAuthError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeAuthError(rec)

		require.Equal(t, api.ApplicationJSON, rec.Header().Get("Content-Type"))
		require.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp api.ErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&errorResp)
		require.NoError(t, err)

		require.Equal(t, api.ErrUnauthorized, errorResp.Error.Code)
		require.Equal(t, "invalid credentials", errorResp.Error.Message)
	})
}
