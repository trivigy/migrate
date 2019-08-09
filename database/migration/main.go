package migration

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/types"
)

// Migration represents a database migration root command.
type Migration struct {
	config map[string]config.Database
}

// NewMigration instantiates a new database migration command and returns it.
func NewMigration(config map[string]config.Database) types.Command {
	return &Migration{config: config}
}

// NewCommand returns a new cobra.Command object.
func (r *Migration) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "Manages the lifecycle of a database migration.",
		Long:         "Manages the lifecycle of a database migration",
		SilenceUsage: true,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		NewGenerate(r.config).(*Generate).NewCommand("generate"),
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
func (r *Migration) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
