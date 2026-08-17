package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var wakeCmd = &cobra.Command{
	Use:   "wake [host]",
	Short: "Wake host using Wake-on-LAN",
	Long:  "Request gateway to send a magic packet to wake a host using Wake-on-LAN",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return &ArgumentNumberError{
				Expected: 1,
				Got:      len(args),
			}
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		hostName := args[0]
		c := client.New("http://"+cfg.Gateway.Address, cfg.Gateway.Secret)

		slog.Debug("requesting gateway to wake host", "host", hostName)
		err := c.Wake(hostName)
		if err != nil {
			slog.Debug("failed to wake the host", "err", err)
			return err
		}

		fmt.Printf("Gateway successfully sent magic packet to %s\n", hostName)

		return nil
	},
}
