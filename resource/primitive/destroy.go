package primitive

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Destroy represents the database destroy command.
type Destroy struct {
	Driver interface {
		types.Destroyer
	} `json:"driver" yaml:"driver"`
}

// DestroyOptions is used for executing the Run() command.
type DestroyOptions struct{}

// NewCommand returns a new cobra.Command destroy command object.
func (r Destroy) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Stops and removes running instance of this resource.",
		Long:  "Stops and removes running instance of this resource",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := DestroyOptions{}
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
func (r Destroy) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Destroy) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the destroy command.
func (r Destroy) Run(out io.Writer, opts DestroyOptions) error {
	if err := r.Driver.Destroy(out); err != nil {
		return err
	}
	return nil
}
