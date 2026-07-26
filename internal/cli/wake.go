package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var wakeCmd = &cobra.Command{
	Use:   "wake [device]",
	Short: "Lol",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%#v\n", cfg)
	},
}
