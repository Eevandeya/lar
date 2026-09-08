package cli

import (
	"io"
	"log/slog"

	"github.com/eevandeya/lar/internal/client"
	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/version"
	"github.com/spf13/cobra"
)

var debug bool

func NewRootCommand(output io.Writer) *cobra.Command {
	var cfg *config.ClientConfig

	cmd := &cobra.Command{
		Use:           "lar [command]",
		Short:         "CLI for managing your machines through a gateway.",
		Version:       version.Version,
		Long:          "Manage your machines through a gateway.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configureLogging(debug)

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
			loaded, err := config.LoadClient(configPath)
			if err != nil {
				slog.Debug("could not load config", "err", err)
				return err
			}

			cfg = loaded

			return nil
		},
	}

	cmd.PersistentFlags().String("config", "", "specify lar config path")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug output")

	cmd.SetOut(output)

	provider := MachineClientProvider{
		getClient: func() MachineClient {
			return client.New(cfg.Gateway.Address, cfg.Gateway.Secret)
		},
	}

	cmd.AddCommand(
		newWakeCommand(output, provider.WakeClient),
		newStatusCommand(output, provider.StatusClient),
		newShutdownCommand(output, provider.ShutdownClient),
	)

	return cmd
}
