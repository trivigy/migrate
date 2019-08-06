package release

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/internal/require"
	"github.com/trivigy/migrate/v2/types"
)

// Delete represents the cluster release delete command object.
type Delete struct {
	config map[string]config.Cluster
}

// NewDelete initializes a new cluster release delete command.
func NewDelete(config map[string]config.Cluster) types.Command {
	return &Delete{config: config}
}

// DeleteOptions is used for executing the Run() command.
type DeleteOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it, and returns it.
func (r *Delete) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Stops a running release and removes all resources.",
		Long:  "Stops a running release and removes all resources",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := DeleteOptions{Env: env}
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
func (r *Delete) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Delete) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the cluster release delete
// command.
func (r *Delete) Run(out io.Writer, opts DeleteOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	if err := cfg.Driver.Setup(out); err != nil {
		return err
	}
	return nil
}
