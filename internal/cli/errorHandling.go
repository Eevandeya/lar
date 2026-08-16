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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isDNSError(err error) bool {
	var dnsError *net.DNSError
	return errors.As(err, &dnsError)
}

func HandleError(cmd *cobra.Command, err error) int {
	errorPrefix := fmt.Sprintf("%sError:%s", red, reset)

	var apiErr client.APIError
	var netErr net.Error
	var argErr *ArgumentNumberError
	var codeErr ExitCodeError
	var flagErr *pflag.NotExistError

	switch {
	case errors.As(err, &apiErr):
		switch apiErr.Code {
		case api.ErrUnauthorized:
			_, _ = fmt.Fprintf(os.Stderr, "%s authentication failed\n", errorPrefix)
		case api.ErrMissingHostName:
			_, _ = fmt.Fprintf(os.Stderr, "%s gateway did not receive host name\n", errorPrefix)
		case api.ErrInvalidHostName:
			var hostErr client.HostError
			if errors.As(err, &hostErr) {
				_, _ = fmt.Fprintf(os.Stderr, "%s host '%s' not found\n", errorPrefix, hostErr.HostName)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "%s host not found\n", errorPrefix)
			}
		case api.ErrInterfaceUnavailable:
			_, _ = fmt.Fprintf(os.Stderr, "%s network interface on gateway is unavailable\n", errorPrefix)
		case api.ErrARPProbingFailed:
			_, _ = fmt.Fprintf(os.Stderr, "%s gateway failed to check host status\n", errorPrefix)
		case api.ErrWOLFailed:
			_, _ = fmt.Fprintf(os.Stderr, "%s gateway failed to send magic packet\n", errorPrefix)
		default:
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorPrefix, apiErr.Message)
		}

	case errors.As(err, &netErr):
		switch {
		case isDNSError(err):
			_, _ = fmt.Fprintf(os.Stderr, "%s failed to resolve gateway\n", errorPrefix)
		case errors.Is(err, syscall.ECONNREFUSED):
			_, _ = fmt.Fprintf(os.Stderr, "%s connection to gateway was refused\n", errorPrefix)
		case isTimeout(err):
			_, _ = fmt.Fprintf(os.Stderr, "%s connection to gateway timed out\n", errorPrefix)
		default:
			_, _ = fmt.Fprintf(os.Stderr, "%s failed to communicate with gateway\n", errorPrefix)
		}

	case errors.As(err, &argErr):
		_, _ = fmt.Fprintf(os.Stderr, "%s '%s' require %d args, but got %d\n\n%s\n",
			errorPrefix, cmd.Name(), argErr.Expected, argErr.Got, cmd.UsageString())

	case errors.Is(err, config.ErrNoConfig):
		_, _ = fmt.Fprintf(os.Stderr, "%s config was not found or passed through --config\n", errorPrefix)

	case errors.As(err, &flagErr):
		_, _ = fmt.Fprintf(os.Stderr, "%s %s\n\n%s\n", errorPrefix, flagErr.Error(), cmd.UsageString())

	case errors.As(err, &codeErr):
		return int(codeErr)

	default:
		_, _ = fmt.Fprintf(os.Stderr, "%s an unexpected error occurred\n", errorPrefix)
	}
	return 1
}
