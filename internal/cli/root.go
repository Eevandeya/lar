package cli

import (
	"log/slog"

	"github.com/eevandeya/lar/internal/config"
	"github.com/spf13/cobra"
)

var cfg *config.ClientConfig
var debug bool

var RootCmd = &cobra.Command{
	Use:           "lar [command]",
	Short:         "CLI for managing your machines through a gateway.",
	Version:       "0.0.1",
	Long:          "Manage your machines through a gateway.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configureLogging(debug)

		if cmd.CommandPath() == "lar version" {
			// TODO: version
			return nil
		}

		slog.Debug("resolving config")
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			slog.Debug("could not get --config flag data", "err", err)
			return err
		}

		if configPath == "" {
			slog.Debug("no config via --config flag, checking default config paths")
			configPath, err = config.LocateClientConfig()
			if err != nil {
				slog.Debug("could not resolve config", "err", err)
				return err
			}
		} else {
			slog.Debug("using config path from --config flag", "path", configPath)
		}

		slog.Debug("loading config", "path", configPath)
		cfg, err = config.LoadClient(configPath)
		if err != nil {
			slog.Debug("could not load config", "err", err)
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().String("config", "", "specify lar config path")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug output")

	RootCmd.AddCommand(wakeCmd)
	RootCmd.AddCommand(shutDownCmd)

	statusCmd.Flags().BoolP("quiet", "q", false, "suppress output")
	RootCmd.AddCommand(statusCmd)
}
