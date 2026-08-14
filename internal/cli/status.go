package cli

import (
	"fmt"

	"github.com/eevandeya/lar/internal/client"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [host]",
	Short: "Lol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostName := args[0]
		c := client.New("http://"+cfg.Gateway.Address, cfg.Gateway.Secret)
		online, err := c.Status(hostName)
		if err != nil {
			return err
		}

		if online {
			fmt.Println("Online")
		} else {
			fmt.Println("Offline")
		}
		return nil
	},
}
