package database

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/internal/enum"
	"github.com/trivigy/migrate/internal/nub"
	"github.com/trivigy/migrate/internal/require"
	"github.com/trivigy/migrate/internal/store"
	"github.com/trivigy/migrate/internal/store/model"
)

// Down represents the database migration down command object.
type Down struct {
	common
	config map[string]Config
}

// NewCreate initializes a new database down command.
func NewDown(config map[string]Config) Command {
	return &Down{config: config}
}

// DownOptions is used for executing the Run() command.
type DownOptions struct {
	Env    string `json:"env" yaml:"env"`
	Num    int    `json:"num" yaml:"num"`
	DryRun bool   `json:"dryRun" yaml:"dryRun"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *Down) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "blah blah blah down.",
		Long:  "blah blah blah down",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			num, err := cmd.Flags().GetInt("num")
			if err != nil {
				return errors.WithStack(err)
			}

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := DownOptions{Env: env, Num: num, DryRun: dryRun}
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
		"num", "n", 1,
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
	config, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	db, err := store.Open(config.Driver.Name(), config.Driver.Source())
	if err != nil {
		return err
	}

	migrationPlan, err := r.GenerateMigrationPlan(db, enum.DirectionDown, config.Migrations)
	if err != nil {
		return err
	}

	steps := len(migrationPlan)
	if opts.Num > 0 && opts.Num <= steps {
		steps = opts.Num
	}

	if opts.DryRun {
		for i := 0; i < steps; i++ {
			fmt.Fprintf(out, "==> migration %q (%s)\n",
				migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
				enum.DirectionDown,
			)
			for _, op := range migrationPlan[i].Down {
				fmt.Fprintf(out, "%s;\n", op.Query)
			}
		}
	} else {
		for i := 0; i < steps; i++ {
			for _, op := range migrationPlan[i].Down {
				err := op.Execute(db, migrationPlan[i], enum.DirectionDown)
				if err != nil {
					return err
				}
			}

			if err := db.Migrations.Delete(&model.Migration{
				Tag: migrationPlan[i].Tag.String(),
			}); err != nil {
				return fmt.Errorf(
					"failed deleting previously applied migration %q (%s)\n",
					migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
					enum.DirectionDown,
				)
			}

			fmt.Fprintf(out, "migration %q successfully removed (%s)\n",
				migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
				enum.DirectionDown,
			)
		}
	}
	return nil
}
