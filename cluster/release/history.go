package release

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/nub"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// History represents the cluster release history command object.
type History struct {
	config map[string]config.Cluster
}

// NewHistory initializes a new cluster release history command.
func NewHistory(config map[string]config.Cluster) types.Command {
	return &History{config: config}
}

// HistoryOptions is used for executing the Run() command.
type HistoryOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *History) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Prints revisions history of deployed releases.",
		Long:  "Prints revisions history of deployed releases",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := HistoryOptions{Env: env}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceUsage: true,
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
func (r *History) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *History) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the cluster release history
// command.
func (r *History) Run(out io.Writer, opts HistoryOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	if err := cfg.Driver.Setup(out); err != nil {
		return err
	}
	return nil
}
