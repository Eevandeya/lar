package config

import (
	"errors"
	"strings"
)

type TLSConfigErrorType int

const (
	CertificatePathMissing TLSConfigErrorType = iota
	PrivateKeyPathMissing
)

type IncompleteTLSConfigErr struct {
	Type TLSConfigErrorType
}

func (e IncompleteTLSConfigErr) Error() string {
	if e.Type == CertificatePathMissing {
		return "invalid tls config: private key path is provided, but certificate path is missing"
	} else if e.Type == PrivateKeyPathMissing {
		return "invalid tls config: certificate path is provided, but private key path is missing"
	}

	return "invalid tls config"
}

type MissingConfigValuesErr []string

func (e MissingConfigValuesErr) Error() string {
	return "missing mandatory values in config: " + strings.Join(e, ", ")
}

var ErrInvalidIPAddress = errors.New("invalid ip address")
var ErrUnsupportedIPVersion = errors.New("only IPv4 is supported")
var ErrInvalidMacAddressFormat = errors.New("invalid MAC address format, only Ethernet MAC (MAC-48) can be used")
