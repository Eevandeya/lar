package client

import (
	"fmt"

	"github.com/eevandeya/lar/internal/api"
)

type APIError struct {
	StatusCode int
	Code       api.ErrorCode
	Message    string
}

func (e APIError) Error() string {
	return e.Message
}

type MachineError struct {
	MachineName string
	Err         error
}

func (e MachineError) Error() string {
	return e.Err.Error()
}

func (e MachineError) Unwrap() error {
	return e.Err
}

type UnexpectedAPIError int

func (e UnexpectedAPIError) Error() string {
	return fmt.Sprintf("unexpected response from gateway (HTTP %d)", e)
}
