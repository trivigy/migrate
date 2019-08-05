package require

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Args validates an args function and appends usage upon error.
func Args(validateArgs func([]string) error) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateArgs(args); err != nil {
			return fmt.Errorf(
				"%s for %q\n\nUsage:  %s",
				err.Error(),
				cmd.CommandPath(),
				cmd.UseLine(),
			)
		}
		return nil
	}
}

// NoArgs checks that no positional arguments were provided.
func NoArgs(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}

// ExactArgs checks that exactly n positional arguments were provided.
func ExactArgs(args []string, n int) error {
	if len(args) != n {
		return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
	}
	return nil
}
