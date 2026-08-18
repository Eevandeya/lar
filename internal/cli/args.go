package cli

import "github.com/spf13/cobra"

func requireOneArg(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return &ArgumentNumberError{
			Expected: 1,
			Got:      len(args),
		}
	}
	return nil
}
