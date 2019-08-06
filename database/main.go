package database

import (
	"io"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/types"
)

const namespace = "database"

// Path generates a key path for a migrations stored in a shared map.
func Path(name string, tag string) string {
	_, caller, _, _ := runtime.Caller(1)
	group := path.Base(path.Dir(caller))
	return strings.Join([]string{namespace, group, tag + "_" + name}, "/")
}

// Filter iterates over all registered database migrations.
func Filter(fn func(migration types.Migration)) types.RangeFunc {
	return func(key, value interface{}) bool {
		fullname := strings.Split(key.(string), "/")
		if fullname[0] == namespace {
			fn(value.(types.Migration))
		}
		return true
	}
}

// Collect iterates over all regirstered cluster migrations and adds them to
// the specified migration.
func Collect(migrations *types.Migrations) types.RangeFunc {
	return Filter(func(migration types.Migration) {
		*migrations = append(*migrations, migration)
	})
}

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
		NewGenerate(r.config).(*Generate).NewCommand("generate"),
		NewCreate(r.config).(*Create).NewCommand("create"),
		NewDestroy(r.config).(*Destroy).NewCommand("destroy"),
		NewUp(r.config).(*Up).NewCommand("up"),
		NewDown(r.config).(*Down).NewCommand("down"),
		NewReport(r.config).(*Report).NewCommand("report"),
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
