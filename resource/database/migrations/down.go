package migrations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/internal/store"
	"github.com/trivigy/migrate/v2/internal/store/model"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Down represents the database migration down command object.
type Down struct {
	Driver interface {
		driver.WithMigrations
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

// downOptions is used for executing the run() command.
type downOptions struct {
	Limit int  `json:"limit" yaml:"limit"`
	Try   bool `json:"dryRun" yaml:"dryRun"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Down)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Down) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name[strings.LastIndex(name, ".")+1:],
		Short: "Rolls back to the previously applied migrations.",
		Long:  "Rolls back to the previously applied migrations",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			if patches, ok := r.Driver.(driver.WithPatches); ok {
				for _, patch := range *patches.Patches(name) {
					if err := patch.Do(ctx, cmd.OutOrStdout()); err != nil {
						return err
					}
				}
			}

			limit, _ := cmd.Flags().GetInt("limit")
			try, _ := cmd.Flags().GetBool("try")
			opts := downOptions{Limit: limit, Try: try}
			return r.run(context.Background(), cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.IntP(
		"limit", "l", 1,
		"Indicate `NUMBER` of migrations to apply. Set `0` for all.",
	)
	flags.Bool(
		"try", false,
		"Simulates and prints resource execution parameters.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Down) Execute(name string, out io.Writer, args []string) error {
	wrap := types.Executor{Name: name, Command: r}
	ctx := context.WithValue(context.Background(), global.RefRoot, wrap)
	cmd := r.NewCommand(ctx, name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Down) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the down command.
func (r Down) run(ctx context.Context, out io.Writer, opts downOptions) error {
	source := bytes.NewBuffer(nil)
	if err := r.Driver.Source(ctx, source); err != nil {
		return err
	}

	u, err := url.Parse(source.String())
	if err != nil {
		return err
	}

	db, err := store.Open(u.Scheme, source.String())
	if err != nil {
		return err
	}

	migrationPlan, err := GenerateMigrationPlan(db, types.DirectionDown, r.Driver.Migrations())
	if err != nil {
		return err
	}

	steps := len(migrationPlan)
	if opts.Limit > 0 && opts.Limit <= steps {
		steps = opts.Limit
	}

	if opts.Try {
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
