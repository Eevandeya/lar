package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var wakeCmd = &cobra.Command{
	Use:   "wake [machine]",
	Short: "Wake machine using Wake-on-LAN",
	Long:  "Request gateway to send a magic packet to wake a machine using Wake-on-LAN",
	Args:  requireOneArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]
		c := client.New("http://"+cfg.Gateway.Address, cfg.Gateway.Secret)

		slog.Debug("requesting gateway to wake machine", "machine", machineName)
		err := c.Wake(machineName)
		if err != nil {
			slog.Debug("failed to wake the machine", "err", err)
			return err
		}

		fmt.Printf("Gateway successfully sent magic packet to %s\n", machineName)

		return nil
	},
}
