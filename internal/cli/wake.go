package cli

import (
	"errors"
	"fmt"
	"net"

	"github.com/eevandeya/lar/internal/wol"
	"github.com/spf13/cobra"
)

var ErrUnknownHost = errors.New("no such host in the config")
var ErrInvalidBroadcast = errors.New("invalid broadcast address")

var wakeCmd = &cobra.Command{
	Use:   "wake [host]",
	Short: "Lol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostName := args[0]
		host, ok := cfg.Hosts[hostName]
		if !ok {
			return ErrUnknownHost
		}

		mac, err := net.ParseMAC(host.MAC)
		if err != nil {
			return err
		}
		broadcast := net.ParseIP(cfg.Broadcast)
		if broadcast == nil {
			return ErrInvalidBroadcast
		}

		err = wol.Wake(mac, broadcast)
		if err != nil {
			return err
		}

		fmt.Printf("Magic packet successfully sent to %s\n", hostName)
		return nil
	},
}
