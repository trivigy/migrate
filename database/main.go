package database

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/database/migration"
	"github.com/trivigy/migrate/v2/nub"
	"github.com/trivigy/migrate/v2/types"
)

// Database represents a database root command.
type Database struct {
	config map[string]config.Database
}

// NewDatabase instantiates a new database command and returns it.
func NewDatabase(config map[string]config.Database) types.Command {
	return &Database{config: config}
}

// NewCommand returns a new cobra.Command object.
func (r *Database) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "SQL database deployment and migrations management tool.",
		Long:         "SQL database deployment and migrations management tool",
		SilenceUsage: true,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		NewCreate(r.config).(*Create).NewCommand("create"),
		NewDestroy(r.config).(*Destroy).NewCommand("destroy"),
		migration.NewMigration(r.config).(*migration.Migration).NewCommand("migration"),
	)

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
func (r *Database) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
