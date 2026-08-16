package cli

import "fmt"

type Command string

type ArgumentNumberError struct {
	Command  Command
	Usage    string
	Expected int
	Got      int
}

func (e *ArgumentNumberError) Error() string {
	return fmt.Sprintf(
		"command %q expects %d argument(s), got %d",
		e.Command,
		e.Expected,
		e.Got,
	)
}

type ExitCodeError int

func (e ExitCodeError) Error() string {
	return ""
}
