// Package database implements the database subcommand structure.
package database

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/resource/database/migrations"
	"github.com/trivigy/migrate/v2/types"
)

// Migrations represents a database migration root command.
type Migrations struct {
	Migrations *types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Migrations)

// NewCommand returns a new cobra.Command object.
func (r Migrations) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " COMMAND",
		Short: "Manages the lifecycle of a database migration.",
		Long:  "Manages the lifecycle of a database migration",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		migrations.Generate{
			Migrations: r.Migrations,
		}.NewCommand("generate"),
		migrations.Up{
			Migrations: r.Migrations,
			Driver:     r.Driver,
		}.NewCommand("up"),
		migrations.Down{
			Migrations: r.Migrations,
			Driver:     r.Driver,
		}.NewCommand("down"),
		migrations.Report{
			Migrations: r.Migrations,
			Driver:     r.Driver,
		}.NewCommand("report"),
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Migrations) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Migrations) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
