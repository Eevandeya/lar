package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

const StatusCommand Command = "status"

var statusCmd = &cobra.Command{
	Use:   "status [host]",
	Short: "Lol",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return &ArgumentNumberError{
				Command:  StatusCommand,
				Usage:    cmd.UsageString(),
				Expected: 1,
				Got:      len(args),
			}
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		hostName := args[0]
		c := client.New("http://"+cfg.Gateway.Address, cfg.Gateway.Secret)

		slog.Debug("requesting status from gateway", "host", hostName)
		online, err := c.Status(hostName)
		if err != nil {
			slog.Debug("fail to obtain host status from gateway", "err", err)
			return err
		}

		if online {
			fmt.Printf("%s: %sonline%s\n", hostName, green, reset)
		} else {
			fmt.Printf("%s: %soffline%s\n", hostName, red, reset)
		}
		return nil
	},
}
