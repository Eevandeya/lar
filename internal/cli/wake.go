package cli

import (
	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var wakeCmd = &cobra.Command{
	Use:   "wake [host]",
	Short: "Lol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostName := args[0]
		c := client.New("http://"+cfg.Gateway.Address, cfg.Gateway.Secret)
		err := c.Wake(hostName)
		return err
	},
}
