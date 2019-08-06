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

// Upgrade represents the cluster release upgrade command object.
type Upgrade struct {
	config map[string]config.Cluster
}

// NewUpgrade initializes a new cluster release upgrade command.
func NewUpgrade(config map[string]config.Cluster) types.Command {
	return &Upgrade{config: config}
}

// UpgradeOptions is used for executing the Run() command.
type UpgradeOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *Upgrade) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Redeploy a modified release and track revision version.",
		Long:  "Redeploy a modified release and track revision version",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := UpgradeOptions{Env: env}
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
func (r *Upgrade) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Upgrade) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the create command.
func (r *Upgrade) Run(out io.Writer, opts UpgradeOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	if err := cfg.Driver.Setup(out); err != nil {
		return err
	}
	return nil
}
