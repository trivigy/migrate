// Package database implements the database subcommand structure.
package database

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/resource/database/migrations"
	"github.com/trivigy/migrate/v2/types"
)

// Migrations represents a database migration root command.
type Migrations struct {
	Migrations *types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		types.Sourced
	} `json:"driver" yaml:"driver"`
}

// NewCommand returns a new cobra.Command object.
func (r Migrations) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "Manages the lifecycle of a database migration.",
		SilenceUsage: true,
	}

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

	pflags := cmd.Flags()
	pflags.Bool("help", false, "Show help information.")
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
