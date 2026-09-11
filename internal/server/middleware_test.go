package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/config"
	"github.com/stretchr/testify/require"
)

const secret = "secret"

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		authHeader *string
		wantStatus int
	}{
		{
			name:       "No auth header",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Empty auth header",
			authHeader: new(""),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "No bearer prefix",
			authHeader: new(secret),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Wrong token",
			authHeader: new("Bearer wrong_token"),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Success",
			authHeader: new("Bearer " + secret),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calledNext bool
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calledNext = true
				w.WriteHeader(http.StatusOK)
			})

			cfg := config.GatewayConfig{
				Server: config.Server{
					Secret: secret,
				},
			}
			s := NewServerWithDependencies(&cfg)

			rec := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
			require.NoError(t, err)
			if tt.authHeader != nil {
				req.Header.Set("Authorization", *tt.authHeader)
			}
			s.authMiddleware(handler).ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatus, rec.Code)

			if rec.Code != http.StatusOK {
				var errResp api.ErrorResponse
				err = json.NewDecoder(rec.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, api.ErrUnauthorized, errResp.Error.Code)
				require.Equal(t, "invalid credentials", errResp.Error.Message)
				require.False(t, calledNext)
			} else {
				require.True(t, calledNext)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		wantStatus    int
	}{
		{
			name: "Explicit status",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Implicit status",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Helper()
				_, err := w.Write([]byte("foobar"))
				require.NoError(t, err)
			}),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		var log struct {
			Method   string
			Path     string
			Status   int
			Duration int64
		}

		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		loggingMiddleware(logger, tt.serverHandler).ServeHTTP(rec, req)

		err := json.NewDecoder(&buf).Decode(&log)
		require.NoError(t, err)

		require.Equal(t, http.MethodGet, log.Method)
		require.Equal(t, "/test", log.Path)
		require.Equal(t, tt.wantStatus, log.Status)
		require.NotZero(t, log.Duration)
	}
}
