package cli

import (
	"io"
	"log/slog"

	"github.com/spf13/cobra"
)

func newWakeCommand(output io.Writer, getClient func() WakeClient) *cobra.Command {
	return &cobra.Command{
		Use:   "wake [machine]",
		Short: "Wake machine using Wake-on-LAN",
		Long:  "Request gateway to send a magic packet to wake a machine using Wake-on-LAN",
		Args:  requireOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			machineName := args[0]
			client := getClient()

			slog.Debug("requesting gateway to wake machine", "machine", machineName)
			err := client.Wake(machineName)
			if err != nil {
				slog.Debug("failed to wake the machine", "err", err)
				return err
			}

			err = printWakeRequestStatus(output, machineName)
			if err != nil {
				slog.Debug("failed to write to output:", "err", err)
			}

			return nil
		},
	}
}
