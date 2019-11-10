package migrations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/internal/store"
	"github.com/trivigy/migrate/v2/internal/store/model"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Report represents the database migration report command object.
type Report struct {
	Migrations *types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Report)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Report) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Prints which migrations were applied and when.",
		Long:  "Prints which migrations were applied and when",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.run(context.Background(), cmd.OutOrStdout())
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Report) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Report) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the up command.
func (r Report) run(ctx context.Context, out io.Writer) error {
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

	if err := db.Migrations.CreateTableIfNotExists(); err != nil {
		return err
	}

	sort.Sort(r.Migrations)
	sortedRegistryMigrations := r.Migrations
	sortedDatabaseMigrations, err := db.Migrations.GetMigrationsSorted()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Tag", "Name", "Applied"})
	table.SetColWidth(60)

	maxSize := max(len(*sortedRegistryMigrations), len(sortedDatabaseMigrations))

	for i := 0; i < maxSize; i++ {
		var rgMig *types.Migration
		if i < len(*sortedRegistryMigrations) {
			rgMig = (*sortedRegistryMigrations)[i]
		}

		var dbMig *model.Migration
		if i < len(sortedDatabaseMigrations) {
			dbMig = &sortedDatabaseMigrations[i]
		}

		if rgMig != nil && dbMig != nil {
			if rgMig.Tag.String() != dbMig.Tag {
				return fmt.Errorf(
					"migration tags mismatch %q != %q",
					rgMig.Tag.String(), dbMig.Tag,
				)
			}

			timestamp := dbMig.Timestamp.Format(time.RFC3339)
			table.Append([]string{dbMig.Tag, dbMig.Name, timestamp})
		} else if rgMig != nil && dbMig == nil {
			table.Append([]string{rgMig.Tag.String(), rgMig.Name, "pending"})
		} else if rgMig == nil && dbMig != nil {
			return fmt.Errorf("migration tags missing %q", dbMig.Tag)
		}
	}

	if len(*r.Migrations) > 0 {
		table.Render()
	}

	return nil
}
