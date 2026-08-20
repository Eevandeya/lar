package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var shutDownCmd = &cobra.Command{
	Use:   "shutdown [machine]",
	Short: "Shutdown machine using ssh",
	Long:  "Request gateway to shutdown a machine using ssh",
	Args:  requireOneArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]
		c := client.New(cfg.Gateway.Address, cfg.Gateway.Secret)

		slog.Debug("requesting gateway to shutdown a machine", "machine", machineName)
		err := c.Shutdown(machineName)
		if err != nil {
			slog.Debug("failed to shutdown the machine", "err", err)
			return err
		}

		fmt.Printf("%s: %sshutdown successful%s\n", machineName, cyan, reset)

		return nil
	},
}
