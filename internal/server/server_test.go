package server

import (
	"testing"

	"github.com/eevandeya/lar/internal/config"
	"github.com/stretchr/testify/require"
)

func TestServerListenAndServeNoTLSConfig(t *testing.T) {
	s := NewServerWithDependencies(&config.GatewayConfig{})

	err := s.ListenAndServe(false)

	require.EqualError(t, err, "no tls config found")
}
