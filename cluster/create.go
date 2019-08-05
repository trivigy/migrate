package cluster

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/internal/nub"
	"github.com/trivigy/migrate/internal/require"
)

// Create represents the cluster create command object.
type Create struct {
	config map[string]Config
}

// NewCreate initializes a new cluster create command.
func NewCreate(config map[string]Config) Command {
	return &Create{config: config}
}

// CreateOptions is used for executing the Run() command.
type CreateOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *Create) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "blah blah blah create.",
		Long:  "blah blah blah create",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := CreateOptions{Env: env}
			return r.Run(cmd.OutOrStdout(), opts)
		},
	}

	pflags := cmd.PersistentFlags()
	pflags.Bool("help", false, "Show help information.")
	pflags.StringP(
		"env", "e", nub.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	return cmd
}

// Execute runs the command.
func (r *Create) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Create) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the create command.
func (r *Create) Run(out io.Writer, opts CreateOptions) error {
	config, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	if err := config.Driver.Setup(out); err != nil {
		return err
	}
	return nil
}
