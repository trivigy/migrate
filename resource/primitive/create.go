// Package primitive implements basic set of commands.
package primitive

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Create represents the database create command object.
type Create struct {
	Driver interface {
		types.Creator
	} `json:"driver" yaml:"driver"`
}

// CreateOptions is used for executing the Run() command.
type CreateOptions struct{}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Create) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Constructs and starts a new instance of this resource.",
		Long:  "Constructs and starts a new instance of this resource",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := CreateOptions{}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceUsage: true,
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Create) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Create) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the create command.
func (r Create) Run(out io.Writer, opts CreateOptions) error {
	if err := r.Driver.Create(out); err != nil {
		return err
	}
	return nil
}
