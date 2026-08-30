package cli

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newShutdownCommand(getClient func() ShutdownClient) *cobra.Command {
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

			fmt.Printf("%s: %sshutdown successful%s\n", machineName, cyan, reset)

			return nil
		},
	}
}
