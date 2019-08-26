package migrations

import (
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/internal/store"
	"github.com/trivigy/migrate/v2/internal/store/model"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Up represents the database up command object.
type Up struct {
	common
	Migrations types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		types.Sourced
	} `json:"driver" yaml:"driver"`
}

// UpOptions is used for executing the Run() command.
type UpOptions struct {
	Limit  int  `json:"limit" yaml:"limit"`
	DryRun bool `json:"dryRun" yaml:"dryRun"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Up) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Executes the next queued migration.",
		Long:  "Executes the next queued migration",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}

			opts := UpOptions{Limit: limit, DryRun: dryRun}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceUsage: true,
	}

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
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Up) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Up) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the up command.
func (r Up) Run(out io.Writer, opts UpOptions) error {
	sort.Sort(r.Migrations)
	source, err := r.Driver.Source()
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

	migrationPlan, err := r.GenerateMigrationPlan(db, types.DirectionUp, r.Migrations)
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
				types.DirectionUp,
			)
			for _, op := range migrationPlan[i].Up {
				fmt.Fprintf(out, "%s;\n", op.Query)
			}
		}
	} else {
		for i := 0; i < steps; i++ {
			for _, op := range migrationPlan[i].Up {
				err := op.Execute(db, migrationPlan[i], types.DirectionUp)
				if err != nil {
					return err
				}
			}

			if err := db.Migrations.Insert(&model.Migration{
				Tag:       migrationPlan[i].Tag.String(),
				Timestamp: time.Now(),
			}); err != nil {
				return fmt.Errorf(
					"failed recording migration %q (%s)",
					migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
					types.DirectionUp,
				)
			}

			fmt.Fprintf(out, "migration %q successfully applied (%s)\n",
				migrationPlan[i].Tag.String()+"_"+migrationPlan[i].Name,
				types.DirectionUp,
			)
		}
	}
	return nil
}
