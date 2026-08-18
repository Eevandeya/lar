package config

import (
	"errors"
	"strings"
)

type MissingConfigValuesErr []string

func (e MissingConfigValuesErr) Error() string {
	return "missing mandatory values in config: " + strings.Join(e, ", ")
}

var ErrInvalidIPAddress = errors.New("invalid ip address")
var ErrUnsupportedIPVersion = errors.New("only IPv4 is supported")
