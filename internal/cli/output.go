package cli

import (
	"fmt"
	"io"
)

const (
	green = "\033[32m"
	red   = "\033[31m"
	cyan  = "\033[36m"
	blue  = "\033[34m"
	reset = "\033[0m"
)

func printStatus(output io.Writer, machine string, online bool) error {
	var err error

	if online {
		_, err = fmt.Fprintf(output, "%s: %sonline%s\n", machine, green, reset)
	} else {
		_, err = fmt.Fprintf(output, "%s: %soffline%s\n", machine, red, reset)
	}

	return err
}

func printShutdownSuccess(output io.Writer, machine string) error {
	_, err := fmt.Fprintf(output, "%s: %sshutdown successful%s\n", machine, cyan, reset)
	return err
}

func printWakeRequestStatus(output io.Writer, machine string) error {
	_, err := fmt.Fprintf(output, "%s: %swake request sent%s\n", machine, blue, reset)
	return err
}
