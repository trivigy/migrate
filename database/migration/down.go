package migration

import (
	"fmt"
	"io"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/store"
	"github.com/trivigy/migrate/v2/internal/store/model"
	"github.com/trivigy/migrate/v2/nub"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Down represents the database migration down command object.
type Down struct {
	common
	config map[string]config.Database
}

// NewDown initializes a new database down command.
func NewDown(config map[string]config.Database) types.Command {
	return &Down{config: config}
}

// DownOptions is used for executing the Run() command.
type DownOptions struct {
	Env    string `json:"env" yaml:"env"`
	Limit  int    `json:"limit" yaml:"limit"`
	DryRun bool   `json:"dryRun" yaml:"dryRun"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *Down) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Rolls back to the previously applied migrations.",
		Long:  "Rolls back to the previously applied migrations",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return errors.WithStack(err)
			}

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := DownOptions{Env: env, Limit: limit, DryRun: dryRun}
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
	flags.IntP(
		"limit", "l", 1,
		"Indicate `NUMBER` of migrations to apply. Set `0` for all.",
	)
	flags.Bool(
		"dry-run", false,
		"Simulate a migration printing planned queries.",
	)
	return cmd
}

// Execute runs the command.
func (r *Down) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Down) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the down command.
func (r *Down) Run(out io.Writer, opts DownOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	source, err := cfg.Driver.Source()
	if err != nil {
		return err
	}

	u, err := url.Parse(source)
	if err != nil {
		return err
	}

	db, err := store.Open(u.Scheme, source)
	if err != nil {
		return err
	}

	migrationPlan, err := r.GenerateMigrationPlan(db, types.DirectionDown, cfg.Migrations)
	if err != nil {
		return err
	}

	steps := len(migrationPlan)
	if opts.Limit > 0 && opts.Limit <= steps {
		steps = opts.Limit
	}

	if opts.DryRun {
		for i := 0; i < steps; i++ {
			fmt.Fprintf(out, "==> migration %q (%s)\n",
				migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
				types.DirectionDown,
			)
			for _, op := range migrationPlan[i].Down {
				fmt.Fprintf(out, "%s;\n", op.Query)
			}
		}
	} else {
		for i := 0; i < steps; i++ {
			for _, op := range migrationPlan[i].Down {
				err := op.Execute(db, migrationPlan[i], types.DirectionDown)
				if err != nil {
					return err
				}
			}

			if err := db.Migrations.Delete(&model.Migration{
				Tag: migrationPlan[i].Tag.String(),
			}); err != nil {
				return fmt.Errorf(
					"failed deleting previously applied migration %q (%s)",
					migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
					types.DirectionDown,
				)
			}

			fmt.Fprintf(out, "migration %q successfully removed (%s)\n",
				migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
				types.DirectionDown,
			)
		}
	}
	return nil
}
