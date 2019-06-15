package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/internal/dao"
	"github.com/trivigy/migrate/internal/enum"
	"github.com/trivigy/migrate/internal/store"
)

type upCommand struct {
	cobra.Command
}

func newUpCommand() *upCommand {
	cmd := &upCommand{}
	cmd.Run = cmd.run
	cmd.Use = "up"
	cmd.Short = "Migrates the database to the most recent version"

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

func (r *upCommand) run(cmd *cobra.Command, args []string) {
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

	migrationPlan := generateMigrationPlan(cmd, db, enum.DirectionUp)
	steps := len(migrationPlan)
	if num > 0 && num <= steps {
		steps = num
	}

	if dryrun {
		for i := 0; i < steps; i++ {
			fmt.Fprintf(cmd.OutOrStdout(),
				"==> migration %q (%s)\n",
				migrationPlan[i].Tag.String(), enum.DirectionUp,
			)
			for _, op := range migrationPlan[i].Up {
				fmt.Fprintf(cmd.OutOrStdout(), "%s;\n", op.Query)
			}
		}
	} else {
		for i := 0; i < steps; i++ {
			for _, op := range migrationPlan[i].Up {
				err := op.Execute(db, migrationPlan[i].Tag, enum.DirectionUp)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), err.Error())
					return
				}
			}

			if err := db.Migrations.Insert(&dao.Migration{
				Tag:       migrationPlan[i].Tag.String(),
				Timestamp: time.Now(),
			}); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(),
					"Error: FAILED recording applied migration %q (%s)\n",
					migrationPlan[i].Tag.String(), enum.DirectionUp,
				)
				return
			}

			fmt.Fprintf(cmd.OutOrStdout(),
				"migration %q successfully applied (%s)\n",
				migrationPlan[i].Tag.String(), enum.DirectionUp,
			)
		}
	}
}
