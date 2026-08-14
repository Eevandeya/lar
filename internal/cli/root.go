package cli

import (
	"github.com/eevandeya/lar/internal/config"
	"github.com/spf13/cobra"
)

var cfg *config.ClientConfig

var RootCmd = &cobra.Command{
	Use:     "lar",
	Short:   "CLI to manage your machines through gateway.",
	Version: "0.0.1",
	Long:    "Booger Aids.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CommandPath() == "lar version" {
			return nil
		}

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		if configPath == "" {
			configPath, err = config.LocateClientConfig()
			if err != nil {
				return err
			}
		}

		cfg, err = config.LoadClient(configPath)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().String("config", "", "specify lar config path")

	RootCmd.AddCommand(wakeCmd)
	RootCmd.AddCommand(shutDownCmd)
	RootCmd.AddCommand(statusCmd)
}
