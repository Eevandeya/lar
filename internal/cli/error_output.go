package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/client"
	"github.com/eevandeya/lar/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var errorPrefix = fmt.Sprintf("%sError:%s", red, reset)
var cobraErrorPrefixes = []string{"unknown command ", "flag needs an argument"}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isCobraError(err error) bool {
	for _, prefix := range cobraErrorPrefixes {
		if strings.HasPrefix(err.Error(), prefix) {
			return true
		}
	}
	return false
}

func handleClientError(output io.Writer, err *client.Error) {
	var apiErr client.APIError
	var unexpectedApiErr client.UnexpectedAPIError
	var dnsErr *net.DNSError

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

	case errors.As(err, &dnsErr):
		_, _ = fmt.Fprintf(output, "%s failed to resolve gateway\n", errorPrefix)

	case errors.Is(err, syscall.ECONNREFUSED):
		_, _ = fmt.Fprintf(output, "%s connection to gateway was refused\n", errorPrefix)

	case isTimeout(err):
		_, _ = fmt.Fprintf(output, "%s connection to gateway timed out\n", errorPrefix)

	default:
		_, _ = fmt.Fprintf(output, "%s client operation failed\n", errorPrefix)
	}
}

func handleConfigError(output io.Writer, err *config.Error) {
	var parseErr *config.ParseError

	switch {
	case errors.Is(err, config.ErrNoConfig):
		_, _ = fmt.Fprintf(output, "%s config was not found or passed through --config\n", errorPrefix)

	case errors.Is(err, config.ErrGatewayNotConfigured):
		_, _ = fmt.Fprintf(output, "%s gateway is not configured\n", errorPrefix)

	case errors.Is(err, os.ErrNotExist):
		if pathErr, ok := errors.AsType[*os.PathError](err); ok {
			_, _ = fmt.Fprintf(output, "%s config was not found at path '%s'\n", errorPrefix, pathErr.Path)
		} else {
			_, _ = fmt.Fprintf(output, "%s config was not found\n", errorPrefix)
		}

	case errors.As(err, &parseErr):
		_, _ = fmt.Fprintf(output, "%s failed to parse config\n", errorPrefix)

	default:
		_, _ = fmt.Fprintf(output, "%s config operation failed\n", errorPrefix)
	}
}

func HandleError(output io.Writer, cmd *cobra.Command, err error) int {
	slog.Debug("formatting error", "err", err)

	var clientErr *client.Error
	var configErr *config.Error
	var exitCodeErr ExitCodeError
	var argErr *ArgumentNumberError
	var flagErr *pflag.NotExistError

	switch {
	case errors.As(err, &clientErr):
		handleClientError(output, clientErr)

	case errors.As(err, &configErr):
		handleConfigError(output, configErr)

	case errors.As(err, &exitCodeErr):
		return int(exitCodeErr)

	case isCobraError(err):
		_, _ = fmt.Fprintf(output, "%s %s\n\n%s\n", errorPrefix, err.Error(), cmd.UsageString())

	case errors.As(err, &argErr):
		_, _ = fmt.Fprintf(output, "%s '%s' require %d args, but got %d\n\n%s\n",
			errorPrefix, cmd.Name(), argErr.Expected, argErr.Got, cmd.UsageString())

	case errors.As(err, &flagErr):
		_, _ = fmt.Fprintf(output, "%s %s\n\n%s\n", errorPrefix, flagErr.Error(), cmd.UsageString())

	default:
		_, _ = fmt.Fprintf(output, "%s an unexpected error occurred\n", errorPrefix)
	}

	return 1
}
