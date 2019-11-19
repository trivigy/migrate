// Package require implements cobra arguments verification helpers.
package require

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Args validates an args function and appends usage upon error.
func Args(validateArgs func(*cobra.Command, []string) error) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateArgs(cmd, args); err != nil {
			return fmt.Errorf(
				"%s for %q\n\n%s",
				err.Error(),
				cmd.CommandPath(),
				cmd.UsageString(),
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

// MaxArgs checks that no more than n positional arguments were provided.
func MaxArgs(args []string, n int) error {
	if len(args) > n {
		return fmt.Errorf("accepts at most %d arg(s), received %d", n, len(args))
	}
	return nil
}