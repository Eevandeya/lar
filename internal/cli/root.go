package cli

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "lar",
	Short: "CLI to manage your machines through gateway.",
	Long:  "Booger Aids.",
}

func init() {
	RootCmd.AddCommand(wakeCmd)
	RootCmd.AddCommand(shutDownCmd)
	RootCmd.AddCommand(statusCmd)
}
