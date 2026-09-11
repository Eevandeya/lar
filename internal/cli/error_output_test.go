package cli

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/client"
	"github.com/eevandeya/lar/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestIsCobraError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantOutput bool
	}{
		{
			name:       "Cobra error",
			err:        errors.New("flag needs an argument: --config"),
			wantOutput: true,
		},
		{
			name:       "Not a cobra error",
			err:        errors.New("foo"),
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantOutput, isCobraError(tt.err))
		})
	}
}

func TestHandleClientError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantOutput string
	}{
		{
			name:       "Connection refused",
			err:        syscall.ECONNREFUSED,
			wantOutput: fmt.Sprintf("%s connection to gateway was refused\n", errorPrefix),
		},
		{
			name: "Unauthorized",
			err: client.APIError{
				Code: api.ErrUnauthorized,
			},
			wantOutput: fmt.Sprintf("%s authentication failed\n", errorPrefix),
		},
		{
			name:       "Missing machine name",
			err:        client.APIError{Code: api.ErrMissingMachineName},
			wantOutput: fmt.Sprintf("%s gateway did not receive machine name\n", errorPrefix),
		},
		{
			name: "Invalid machine name",
			err: client.MachineError{
				MachineName: "pc1",
				Err:         client.APIError{Code: api.ErrInvalidMachineName},
			},
			wantOutput: fmt.Sprintf("%s machine 'pc1' not found\n", errorPrefix),
		},
		{
			name:       "Invalid machine name without machine details",
			err:        client.APIError{Code: api.ErrInvalidMachineName},
			wantOutput: fmt.Sprintf("%s machine not found\n", errorPrefix),
		},
		{
			name:       "Interface unavailable",
			err:        client.APIError{Code: api.ErrInterfaceUnavailable},
			wantOutput: fmt.Sprintf("%s network interface on gateway is unavailable\n", errorPrefix),
		},
		{
			name:       "ARP probing failed",
			err:        client.APIError{Code: api.ErrARPProbingFailed},
			wantOutput: fmt.Sprintf("%s gateway failed to check machine status\n", errorPrefix),
		},
		{
			name:       "WOL failed",
			err:        client.APIError{Code: api.ErrWOLFailed},
			wantOutput: fmt.Sprintf("%s gateway failed to send magic packet\n", errorPrefix),
		},
		{
			name:       "SSH failed",
			err:        client.APIError{Code: api.ErrSSHFailed},
			wantOutput: fmt.Sprintf("%s gateway failed to shutdown machine via ssh\n", errorPrefix),
		},
		{
			name:       "Unknown API error",
			err:        client.APIError{Message: "gateway failure"},
			wantOutput: fmt.Sprintf("%s gateway failure\n", errorPrefix),
		},
		{
			name:       "Unexpected API response",
			err:        client.UnexpectedAPIError(502),
			wantOutput: fmt.Sprintf("%s unexpected response from gateway (HTTP 502)\n", errorPrefix),
		},
		{
			name:       "DNS resolution failed",
			err:        &net.DNSError{Name: "gateway.example.com"},
			wantOutput: fmt.Sprintf("%s failed to resolve gateway\n", errorPrefix),
		},
		{
			name:       "Connection timed out",
			err:        timeoutError{},
			wantOutput: fmt.Sprintf("%s connection to gateway timed out\n", errorPrefix),
		},
		{
			name:       "Unknown client error",
			err:        errors.New("unexpected client failure"),
			wantOutput: fmt.Sprintf("%s client operation failed\n", errorPrefix),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			handleClientError(&output, &client.Error{Err: tt.err})

			require.Equal(t, tt.wantOutput, output.String())
		})
	}
}

func TestHandleConfigError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantOutput string
	}{
		{
			name:       "Config was not found",
			err:        config.ErrNoConfig,
			wantOutput: fmt.Sprintf("%s config was not found or passed through --config\n", errorPrefix),
		},
		{
			name:       "Gateway is not configured",
			err:        config.ErrGatewayNotConfigured,
			wantOutput: fmt.Sprintf("%s gateway is not configured\n", errorPrefix),
		},
		{
			name: "Config path was not found",
			err: &os.PathError{
				Op:   "open",
				Path: "/tmp/config.yaml",
				Err:  os.ErrNotExist,
			},
			wantOutput: fmt.Sprintf("%s config was not found at path '/tmp/config.yaml'\n", errorPrefix),
		},
		{
			name:       "Config was not found without path",
			err:        os.ErrNotExist,
			wantOutput: fmt.Sprintf("%s config was not found\n", errorPrefix),
		},
		{
			name: "Config parsing failed",
			err: &config.ParseError{
				Err: errors.New("invalid yaml"),
			},
			wantOutput: fmt.Sprintf("%s failed to parse config\n", errorPrefix),
		},
		{
			name:       "Unknown config error",
			err:        errors.New("unexpected config failure"),
			wantOutput: fmt.Sprintf("%s config operation failed\n", errorPrefix),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			handleConfigError(&output, &config.Error{Err: tt.err})

			require.Equal(t, tt.wantOutput, output.String())
		})
	}
}

func TestHandleError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	_, flagErr := cmd.Flags().GetString("missing")
	require.Error(t, flagErr)

	usage := cmd.UsageString()

	tests := []struct {
		name          string
		err           error
		wantIntOutput int
		wantOutput    string
	}{
		{
			name: "Client error",
			err: &client.Error{
				Err: client.APIError{Code: api.ErrUnauthorized},
			},
			wantIntOutput: 1,
			wantOutput:    fmt.Sprintf("%s authentication failed\n", errorPrefix),
		},
		{
			name:          "Config error",
			err:           &config.Error{Err: config.ErrNoConfig},
			wantIntOutput: 1,
			wantOutput: fmt.Sprintf(
				"%s config was not found or passed through --config\n",
				errorPrefix,
			),
		},
		{
			name:          "Exit code error",
			err:           ExitCodeError(7),
			wantIntOutput: 7,
			wantOutput:    "",
		},
		{
			name:          "Unknown Cobra command",
			err:           errors.New("unknown command wake"),
			wantIntOutput: 1,
			wantOutput:    fmt.Sprintf("%s unknown command wake\n\n%s\n", errorPrefix, usage),
		},
		{
			name:          "Cobra flag argument is missing",
			err:           errors.New("flag needs an argument: --config"),
			wantIntOutput: 1,
			wantOutput:    fmt.Sprintf("%s flag needs an argument: --config\n\n%s\n", errorPrefix, usage),
		},
		{
			name: "Argument count error",
			err: &ArgumentNumberError{
				Expected: 2,
				Got:      1,
			},
			wantIntOutput: 1,
			wantOutput: fmt.Sprintf(
				"%s '%s' require 2 args, but got 1\n\n%s\n",
				errorPrefix,
				cmd.Name(),
				usage,
			),
		},
		{
			name:          "Flag does not exist",
			err:           flagErr,
			wantIntOutput: 1,
			wantOutput:    fmt.Sprintf("%s %s\n\n%s\n", errorPrefix, flagErr.Error(), usage),
		},
		{
			name:          "Unknown error",
			err:           errors.New("unexpected error"),
			wantIntOutput: 1,
			wantOutput:    fmt.Sprintf("%s an unexpected error occurred\n", errorPrefix),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			gotCode := HandleError(&output, cmd, tt.err)

			require.Equal(t, tt.wantIntOutput, gotCode)
			require.Equal(t, tt.wantOutput, output.String())
		})
	}
}
