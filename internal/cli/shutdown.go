package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var shutDownCmd = &cobra.Command{
	Use:   "shutdown [host]",
	Short: "Shutdown host using ssh",
	Long:  "Request gateway to shutdown a host using ssh",
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

		slog.Debug("requesting gateway to shutdown a host", "host", hostName)
		err := c.Shutdown(hostName)
		if err != nil {
			slog.Debug("failed to shutdown the host", "err", err)
			return err
		}

		fmt.Printf("Gateway successfully shutdown %s\n", hostName)

		return nil
	},
}
