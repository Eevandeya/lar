package cli

import (
	"fmt"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [machine]",
	Short: "Check machine status",
	Long:  "Check whether a machine is online.",
	Args:  requireOneArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]
		c := client.New("http://"+cfg.Gateway.Address, cfg.Gateway.Secret)

		slog.Debug("requesting status from gateway", "machine", machineName)
		online, err := c.Status(machineName)
		if err != nil {
			slog.Debug("failed to obtain machine status from gateway", "err", err)
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
			fmt.Printf("%s: %sonline%s\n", machineName, green, reset)
		} else {
			fmt.Printf("%s: %soffline%s\n", machineName, red, reset)
		}
		return nil
	},
}
