package main

import (
	"os"

	"github.com/eevandeya/lar/internal/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		os.Exit(cli.HandleError(err))
	}
}
