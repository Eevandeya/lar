package config

import (
	"strings"
)

type TLSConfigErrorType int

const (
	CertificatePathMissing TLSConfigErrorType = iota
	PrivateKeyPathMissing
)

type Error struct {
	Err error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

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
	if len(e) < 1 {
		return "missing mandatory values in config"
	}
	return "missing mandatory values in config: " + strings.Join(e, ", ")
}

type ParseError struct {
	Err error
}

func (e *ParseError) Error() string {
	return e.Err.Error()
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
