// Package resource implements pluggable cobra subcommands.
package resource

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Collection represents a collection aggregator command.
type Environment map[string]types.Resource

var _ interface {
	types.Resource
	types.Command
} = new(Environment)

// NewCommand returns a new cobra.Command object.
func (r Environment) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:  name[strings.LastIndex(name, ".")+1:] + " COMMAND",
		Args: require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	for key, resource := range r {
		cmd.AddCommand(resource.NewCommand(ctx, name+"."+key))
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Environment) Execute(name string, out io.Writer, args []string) error {
	wrap := types.Executor{Name: name, Command: r}
	ctx := context.WithValue(context.Background(), global.RefRoot, wrap)
	cmd := r.NewCommand(ctx, name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Environment) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
