package api

type ErrorCode string

const (
	ErrUnauthorized         ErrorCode = "unauthorized"
	ErrMissingHostName      ErrorCode = "missing_host_name"
	ErrInvalidHostName      ErrorCode = "invalid_host_name"
	ErrInterfaceUnavailable ErrorCode = "interface_unavailable"
	ErrARPProbingFailed     ErrorCode = "arp_probing_failed"
	ErrWOLFailed            ErrorCode = "wol_failed"
	ErrSSHFailed            ErrorCode = "ssh_failed"
)
