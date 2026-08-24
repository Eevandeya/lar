package client

import (
	"net/http"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/stretchr/testify/require"
)

func TestAPIErrorError(t *testing.T) {
	message := "invalid machine name"
	err := APIError{
		StatusCode: http.StatusBadRequest,
		Code:       api.ErrInvalidMachineName,
		Message:    message,
	}

	require.Equal(t, message, err.Error())
}

func TestMachineErrorError(t *testing.T) {
	message := "invalid machine name"
	err := MachineError{
		MachineName: "pc1",
		Err: APIError{
			StatusCode: http.StatusBadGateway,
			Code:       api.ErrInvalidMachineName,
			Message:    message,
		}}
	require.Equal(t, message, err.Error())
}

func TestMachineErrorUnwrap(t *testing.T) {
	innerErr := APIError{
		StatusCode: http.StatusBadGateway,
		Code:       api.ErrInvalidMachineName,
		Message:    "invalid machine name",
	}

	err := MachineError{
		MachineName: "pc1",
		Err:         innerErr,
	}
	require.Equal(t, innerErr, err.Unwrap())
}

func TestUnexpectedAPIErrorError(t *testing.T) {
	err := UnexpectedAPIError(http.StatusBadRequest)
	require.Equal(t, "unexpected response from gateway (HTTP 400)", err.Error())
}
