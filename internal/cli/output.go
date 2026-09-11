package cli

import (
	"fmt"
	"io"
)

const (
	Green = "\033[32m"
	Red   = "\033[31m"
	Cyan  = "\033[36m"
	Blue  = "\033[34m"
	Reset = "\033[0m"
)

func printStatus(output io.Writer, machine string, online bool) error {
	var err error

	if online {
		_, err = fmt.Fprintf(output, "%s: %sonline%s\n", machine, Green, Reset)
	} else {
		_, err = fmt.Fprintf(output, "%s: %soffline%s\n", machine, Red, Reset)
	}

	return err
}

func printShutdownSuccess(output io.Writer, machine string) error {
	_, err := fmt.Fprintf(output, "%s: %sshutdown successful%s\n", machine, Cyan, Reset)
	return err
}

func printWakeRequestStatus(output io.Writer, machine string) error {
	_, err := fmt.Fprintf(output, "%s: %swake request sent%s\n", machine, Blue, Reset)
	return err
}
