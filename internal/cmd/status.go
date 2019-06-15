package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/internal/dao"
	"github.com/trivigy/migrate/internal/dto"
	"github.com/trivigy/migrate/internal/store"
)

type statusCommand struct {
	cobra.Command
}

func newStatusCommand() *statusCommand {
	cmd := &statusCommand{}
	cmd.Run = cmd.run
	cmd.Use = "status"
	cmd.Short = "Show migration status for the current database"

	flags := cmd.Flags()
	flags.SortFlags = false
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

func (r *statusCommand) run(cmd *cobra.Command, args []string) {
	env, _ := cmd.Flags().GetString("env")
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

	if err := db.Migrations.CreateTableIfNotExists(); err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", err)
		return
	}

	sortedRegistryMigrations := getSortedRegistryMigrations()
	sortedDatabaseMigrations, err := db.Migrations.GetMigrationsSorted()
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", err)
		return
	}

	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader([]string{"Tag", "Applied"})
	table.SetColWidth(60)

	maxSize := max(len(sortedRegistryMigrations), len(sortedDatabaseMigrations))

	for i := 0; i < maxSize; i++ {
		var rgMig *dto.Migration
		if i < len(sortedRegistryMigrations) {
			rgMig = &sortedRegistryMigrations[i]
		}

		var dbMig *dao.Migration
		if i < len(sortedDatabaseMigrations) {
			dbMig = &sortedDatabaseMigrations[i]
		}

		if rgMig != nil && dbMig != nil {
			if rgMig.Tag.String() != dbMig.Tag {
				fmt.Fprintf(cmd.OutOrStderr(),
					"error: migration tags mismatch %q != %q\n",
					rgMig.Tag.String(), dbMig.Tag,
				)
				return
			}

			timestamp := dbMig.Timestamp.Format(time.RFC3339)
			table.Append([]string{dbMig.Tag, timestamp})
		} else if rgMig != nil && dbMig == nil {
			table.Append([]string{rgMig.Tag.String(), "pending"})
		} else if rgMig == nil && dbMig != nil {
			fmt.Fprintf(cmd.OutOrStderr(),
				"error: migration tags missing %q\n", dbMig.Tag,
			)
			return
		}

	}
	table.Render()
}
