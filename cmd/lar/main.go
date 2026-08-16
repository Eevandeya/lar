package main

import (
	"os"

	"github.com/eevandeya/lar/internal/cli"
)

func main() {
	cmd, err := cli.RootCmd.ExecuteC()
	if err != nil {
		os.Exit(cli.HandleError(cmd, err))
	}
}
