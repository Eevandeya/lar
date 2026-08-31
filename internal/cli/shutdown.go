package cli

import (
	"io"
	"log/slog"

	"github.com/spf13/cobra"
)

func newShutdownCommand(output io.Writer, getClient func() ShutdownClient) *cobra.Command {
	return &cobra.Command{
		Use:   "shutdown [machine]",
		Short: "Shutdown machine using ssh",
		Long:  "Request gateway to shutdown a machine using ssh",
		Args:  requireOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			machineName := args[0]
			client := getClient()

			slog.Debug("requesting gateway to shutdown a machine", "machine", machineName)
			err := client.Shutdown(machineName)
			if err != nil {
				slog.Debug("failed to shutdown the machine", "err", err)
				return err
			}

			err = printShutdownSuccess(output, machineName)
			if err != nil {
				slog.Debug("failed to write to output:", "err", err)
			}

			return nil
		},
	}
}
