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

type HostError struct {
	HostName string
	Err      error
}

func (e HostError) Error() string {
	return e.Err.Error()
}

func (e HostError) Unwrap() error {
	return e.Err
}

type UnexpectedAPIError int

func (e UnexpectedAPIError) Error() string {
	return fmt.Sprintf("unexpected response from gateway (HTTP %d)", e)
}
