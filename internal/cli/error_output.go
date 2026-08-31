package cli

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
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

func isUnknownCommand(err error) bool {
	return strings.HasPrefix(err.Error(), "unknown command ")
}

func HandleError(output io.Writer, cmd *cobra.Command, err error) int {
	errorPrefix := fmt.Sprintf("%sError:%s", red, reset)

	var apiErr client.APIError
	var unexpectedApiErr client.UnexpectedAPIError
	var netErr net.Error
	var argErr *ArgumentNumberError
	var codeErr ExitCodeError
	var flagErr *pflag.NotExistError

	switch {
	case errors.As(err, &apiErr):
		switch apiErr.Code {
		case api.ErrUnauthorized:
			_, _ = fmt.Fprintf(output, "%s authentication failed\n", errorPrefix)
		case api.ErrMissingMachineName:
			_, _ = fmt.Fprintf(output, "%s gateway did not receive machine name\n", errorPrefix)
		case api.ErrInvalidMachineName:
			if machineErr, ok := errors.AsType[client.MachineError](err); ok {
				_, _ = fmt.Fprintf(output, "%s machine '%s' not found\n", errorPrefix, machineErr.MachineName)
			} else {
				_, _ = fmt.Fprintf(output, "%s machine not found\n", errorPrefix)
			}
		case api.ErrInterfaceUnavailable:
			_, _ = fmt.Fprintf(output, "%s network interface on gateway is unavailable\n", errorPrefix)
		case api.ErrARPProbingFailed:
			_, _ = fmt.Fprintf(output, "%s gateway failed to check machine status\n", errorPrefix)
		case api.ErrWOLFailed:
			_, _ = fmt.Fprintf(output, "%s gateway failed to send magic packet\n", errorPrefix)
		case api.ErrSSHFailed:
			_, _ = fmt.Fprintf(output, "%s gateway failed to shutdown machine via ssh\n", errorPrefix)
		default:
			_, _ = fmt.Fprintf(output, "%s %s\n", errorPrefix, apiErr.Message)
		}

	case errors.As(err, &unexpectedApiErr):
		_, _ = fmt.Fprintf(output, "%s unexpected response from gateway (HTTP %d)\n", errorPrefix, unexpectedApiErr)

	case errors.As(err, &netErr):
		switch {
		case isDNSError(err):
			_, _ = fmt.Fprintf(output, "%s failed to resolve gateway\n", errorPrefix)
		case errors.Is(err, syscall.ECONNREFUSED):
			_, _ = fmt.Fprintf(output, "%s connection to gateway was refused\n", errorPrefix)
		case isTimeout(err):
			_, _ = fmt.Fprintf(output, "%s connection to gateway timed out\n", errorPrefix)
		default:
			_, _ = fmt.Fprintf(output, "%s failed to communicate with gateway\n", errorPrefix)
		}

	case errors.As(err, &argErr):
		_, _ = fmt.Fprintf(output, "%s '%s' require %d args, but got %d\n\n%s\n",
			errorPrefix, cmd.Name(), argErr.Expected, argErr.Got, cmd.UsageString())

	case errors.Is(err, config.ErrNoConfig):
		_, _ = fmt.Fprintf(output, "%s config was not found or passed through --config\n", errorPrefix)

	case errors.As(err, &flagErr):
		_, _ = fmt.Fprintf(output, "%s %s\n\n%s\n", errorPrefix, flagErr.Error(), cmd.UsageString())

	case isUnknownCommand(err):
		_, _ = fmt.Fprintf(output, "%s %s\n\n%s\n", errorPrefix, err.Error(), cmd.UsageString())

	case errors.As(err, &codeErr):
		return int(codeErr)

	default:
		_, _ = fmt.Fprintf(output, "%s an unexpected error occurred\n", errorPrefix)
	}
	return 1
}
