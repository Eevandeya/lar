package cli

import (
	"io"
	"log/slog"

	"github.com/spf13/cobra"
)

func newStatusCommand(output io.Writer, getClient func() StatusClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [machine]",
		Short: "Check machine status",
		Long:  "Check whether a machine is online.",
		Args:  requireOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			machineName := args[0]
			client := getClient()

			slog.Debug("requesting status from gateway", "machine", machineName)
			online, err := client.Status(machineName)
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

			err = printStatus(output, machineName, online)
			if err != nil {
				slog.Debug("failed to write to output:", "err", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "suppress output")

	return cmd
}
