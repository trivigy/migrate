// Package resource implements pluggable cobra subcommands.
package resource

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Collection represents a collection aggregator command.
type Collection map[string]types.Resource

// NewCommand returns a new cobra.Command object.
func (r Collection) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:  name + " COMMAND",
		Args: require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	for name, resource := range r {
		cmd.AddCommand(resource.NewCommand(name))
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Collection) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Collection) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
