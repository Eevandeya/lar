package main

import (
	"os"

	"github.com/eevandeya/lar/internal/cli"
)

func main() {
	cmd, err := cli.NewRootCommand(os.Stdout).ExecuteC()
	if err != nil {
		os.Exit(cli.HandleError(os.Stderr, cmd, err))
	}
}
