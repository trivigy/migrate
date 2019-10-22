package database

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Source represents the database source command object.
type Source struct {
	Driver interface {
		types.Sourced
	} `json:"driver" yaml:"driver"`
}

// SourceOptions is used for executing the Run() command.
type SourceOptions struct{}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Source) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Print the data source name as a connection string.",
		Long:  "Print the data source name as a connection string",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := SourceOptions{}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceUsage: true,
	}

	pflags := cmd.PersistentFlags()
	pflags.Bool("help", false, "Show help information.")
	pflags.StringP(
		"env", "e", global.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	return cmd
}

// Execute runs the command.
func (r Source) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Source) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the source command.
func (r Source) Run(out io.Writer, opts SourceOptions) error {
	source, err := r.Driver.Source()
	if err != nil {
		return err
	}

	if _, err := out.Write([]byte(source + "\n")); err != nil {
		return err
	}
	return nil
}
