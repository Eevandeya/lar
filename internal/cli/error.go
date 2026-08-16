package cli

import "fmt"

type Command string

type ArgumentNumberError struct {
	Expected int
	Got      int
}

func (e *ArgumentNumberError) Error() string {
	return fmt.Sprintf("expected %d argument(s), but got %d", e.Expected, e.Got)
}

type ExitCodeError int

func (e ExitCodeError) Error() string {
	return ""
}
