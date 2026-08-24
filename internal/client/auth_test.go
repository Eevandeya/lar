package client

import (
	"net/http"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/stretchr/testify/require"
)

func TestSetAuthHeader(t *testing.T) {
	secret := "secret"
	client := New("https://example.com", secret)
	req, err := http.NewRequest(http.MethodGet, client.baseURL, nil)
	require.NoError(t, err)

	client.setAuthHeader(req)
	require.Equal(t, api.BearerPrefix+secret, req.Header.Get("Authorization"))
}
