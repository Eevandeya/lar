package api

type ErrorCode string

const (
	ErrUnauthorized         ErrorCode = "unauthorized"
	ErrMissingMachineName   ErrorCode = "missing_machine_name"
	ErrInvalidMachineName   ErrorCode = "invalid_machine_name"
	ErrInterfaceUnavailable ErrorCode = "interface_unavailable"
	ErrARPProbingFailed     ErrorCode = "arp_probing_failed"
	ErrWOLFailed            ErrorCode = "wol_failed"
	ErrSSHFailed            ErrorCode = "ssh_failed"
)
