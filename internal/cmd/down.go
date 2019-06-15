package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/internal/dao"
	"github.com/trivigy/migrate/internal/enum"
	"github.com/trivigy/migrate/internal/store"
)

type downCommand struct {
	cobra.Command
}

func newDownCommand() *downCommand {
	cmd := &downCommand{}
	cmd.Run = cmd.run
	cmd.Use = "down"
	cmd.Short = "Undo the last applied database migration"

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool(
		"dry-run", false,
		"Simulate a migration printing planned queries.",
	)
	flags.IntP(
		"num", "n", 0,
		"Indicate `NUMBER` of migrations to apply.",
	)
	flags.StringP(
		"env", "e", "",
		"Run with configurations named `ENV`. (required)",
	)
	flags.Bool("help", false, "Show help information.")

	if err := cmd.MarkFlagRequired("env"); err != nil {
		fmt.Printf("%+v\n", errors.WithStack(err))
		os.Exit(1)
	}
	return cmd
}

func (r *downCommand) run(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("env")
	num, _ := cmd.Flags().GetInt("num")
	dryrun, _ := cmd.Flags().GetBool("dry-run")
	driver := configs.GetString(env + ".driver")
	source := configs.GetString(env + ".source")

	if db == nil {
		var err error
		db, err = store.Open(driver, source)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
			return
		}
		defer db.Close()
	}

	migrationPlan := generateMigrationPlan(cmd, db, enum.DirectionDown)
	steps := len(migrationPlan)
	if num > 0 && num <= steps {
		steps = num
	}

	if dryrun {
		for i := 0; i < steps; i++ {
			fmt.Fprintf(cmd.OutOrStdout(),
				"==> migration %q (%s)\n",
				migrationPlan[i].Tag.String(), enum.DirectionDown,
			)
			for _, op := range migrationPlan[i].Down {
				fmt.Fprintf(cmd.OutOrStdout(), "%s;\n", op.Query)
			}
		}
	} else {
		for i := 0; i < steps; i++ {
			for _, op := range migrationPlan[i].Down {
				err := op.Execute(db, migrationPlan[i].Tag, enum.DirectionDown)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), err.Error())
					return
				}
			}

			if err := db.Migrations.Delete(&dao.Migration{
				Tag: migrationPlan[i].Tag.String(),
			}); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(),
					"error: failed deleting previously applied migration %q (%s)\n",
					migrationPlan[i].Tag.String(), enum.DirectionDown,
				)
				return
			}

			fmt.Fprintf(cmd.OutOrStdout(),
				"migration %q successfully removed (%s)\n",
				migrationPlan[i].Tag.String(), enum.DirectionDown,
			)
		}
	}
}
