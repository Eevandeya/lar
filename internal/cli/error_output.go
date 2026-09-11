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

var errorPrefix = fmt.Sprintf("%sError:%s", Red, Reset)
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

func printError(output io.Writer, msg string) {
	_, _ = fmt.Fprintf(output, "%s %s\n", errorPrefix, msg)
}

func handleClientError(output io.Writer, err *client.Error) {
	var apiErr client.APIError
	var unexpectedApiErr client.UnexpectedAPIError
	var dnsErr *net.DNSError

	switch {
	case errors.As(err, &apiErr):
		switch apiErr.Code {
		case api.ErrUnauthorized:
			printError(output, "authentication failed")
		case api.ErrMissingMachineName:
			printError(output, "gateway did not receive machine name")
		case api.ErrInvalidMachineName:
			if machineErr, ok := errors.AsType[client.MachineError](err); ok {
				printError(output, fmt.Sprintf("machine '%s' not found", machineErr.MachineName))
			} else {
				printError(output, "machine not found")
			}
		case api.ErrInterfaceUnavailable:
			printError(output, "network interface on gateway is unavailable")
		case api.ErrARPProbingFailed:
			printError(output, "gateway failed to check machine status")
		case api.ErrWOLFailed:
			printError(output, "gateway failed to send magic packet")
		case api.ErrSSHFailed:
			printError(output, "gateway failed to shutdown machine via ssh")
		default:
			printError(output, apiErr.Message)
		}

	case errors.As(err, &unexpectedApiErr):
		printError(output, fmt.Sprintf("unexpected response from gateway (HTTP %d)", unexpectedApiErr))

	case errors.As(err, &dnsErr):
		printError(output, "failed to resolve gateway")

	case errors.Is(err, syscall.ECONNREFUSED):
		printError(output, "connection to gateway was refused")

	case errors.Is(err, syscall.ECONNRESET):
		printError(output, "gateway closed the connection")

	case isTimeout(err):
		printError(output, "connection to gateway timed out")

	default:
		printError(output, "client operation failed")
	}
}

func handleConfigError(output io.Writer, err *config.Error) {
	var parseErr *config.ParseError

	switch {
	case errors.Is(err, config.ErrNoConfig):
		printError(output, "config was not found or passed through --config")

	case errors.Is(err, config.ErrGatewayNotConfigured):
		printError(output, "gateway is not configured")

	case errors.Is(err, os.ErrNotExist):
		if pathErr, ok := errors.AsType[*os.PathError](err); ok {
			printError(output, fmt.Sprintf("config was not found at path '%s'", pathErr.Path))
		} else {
			printError(output, "config was not found")
		}

	case errors.As(err, &parseErr):
		printError(output, "failed to parse config")

	default:
		printError(output, "config operation failed")
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
		printError(output, fmt.Sprintf("%s\n\n%s", err.Error(), cmd.UsageString()))

	case errors.As(err, &argErr):
		printError(output, fmt.Sprintf("'%s' require %d args, but got %d\n\n%s",
			cmd.Name(), argErr.Expected, argErr.Got, cmd.UsageString()))

	case errors.As(err, &flagErr):
		printError(output, fmt.Sprintf("%s\n\n%s", flagErr.Error(), cmd.UsageString()))

	default:
		printError(output, "an unexpected error occurred")
	}

	return 1
}
