package cli

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/client"
	"github.com/eevandeya/lar/internal/config"
)

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isDNSError(err error) bool {
	var dnsError *net.DNSError
	return errors.As(err, &dnsError)
}

func HandleError(err error) int {
	var apiErr client.APIError
	var netErr net.Error
	var argErr *ArgumentNumberError
	var codeErr ExitCodeError

	switch {
	case errors.As(err, &apiErr):
		switch apiErr.Code {
		case api.ErrUnauthorized:
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s authentication failed\n", red, reset)
		case api.ErrMissingHostName:
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s gateway did not receive host name\n", red, reset)
		case api.ErrInvalidHostName:
			var hostErr client.HostError
			if errors.As(err, &hostErr) {
				_, _ = fmt.Fprintf(os.Stderr, "%sError:%s host '%s' not found\n", red, reset, hostErr.HostName)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "%sError:%s host not found\n", red, reset)
			}
		case api.ErrInterfaceUnavailable:
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s network interface on gateway is unavailable\n", red, reset)
		case api.ErrARPProbingFailed:
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s failed to check host status\n", red, reset)
		default:
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s %s\n", red, reset, apiErr.Message)
		}

	case errors.As(err, &netErr):
		switch {
		case isDNSError(err):
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s failed to resolve gateway\n", red, reset)
		case errors.Is(err, syscall.ECONNREFUSED):
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s connection to gateway was refused\n", red, reset)
		case isTimeout(err):
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s connection to gateway timed out\n", red, reset)
		default:
			_, _ = fmt.Fprintf(os.Stderr, "%sError:%s failed to communicate with gateway\n", red, reset)
		}

	case errors.As(err, &argErr):
		_, _ = fmt.Fprintf(os.Stderr, "%sError:%s '%s' require %d args, but got %d\n\n%s",
			red, reset, argErr.Command, argErr.Expected, argErr.Got, argErr.Usage)

	case errors.Is(err, config.ErrNoConfig):
		_, _ = fmt.Fprintf(os.Stderr, "%sError:%s config was not found or passed through --config\n", red, reset)

	case errors.As(err, &codeErr):
		return int(codeErr)

	default:
		_, _ = fmt.Fprintf(os.Stderr, "%sError:%s an unexpected error occurred\n", red, reset)
	}
	return 1
}
