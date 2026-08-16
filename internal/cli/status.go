package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [host]",
	Short: "Check host status",
	Long:  "Check whether a host is online.",
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

		slog.Debug("requesting status from gateway", "host", hostName)
		online, err := c.Status(hostName)
		if err != nil {
			slog.Debug("fail to obtain host status from gateway", "err", err)
			return err
		}

		quiet, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			slog.Debug("failed to read --quiet flag", "err", err)
			return err
		}

		if quiet && online {
			return ExitCodeError(0)
		}
		if quiet && !online {
			return ExitCodeError(1)
		}

		if online {
			fmt.Printf("%s: %sonline%s\n", hostName, green, reset)
		} else {
			fmt.Printf("%s: %soffline%s\n", hostName, red, reset)
		}
		return nil
	},
}
