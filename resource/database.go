package resource

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/resource/database"
	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

// Database represents a database root command.
type Database struct {
	Migrations *types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		types.Creator
		types.Destroyer
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Database)

// NewCommand returns a new cobra.Command object.
func (r Database) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " COMMAND",
		Short: "SQL database deployment and migrations management tool.",
		Long:  "SQL database deployment and migrations management tool",
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
		primitive.Create{
			Driver: r.Driver,
		}.NewCommand("create"),
		primitive.Destroy{
			Driver: r.Driver,
		}.NewCommand("destroy"),
		primitive.Source{
			Driver: r.Driver,
		}.NewCommand("source"),
		database.Migrations{
			Migrations: r.Migrations,
			Driver:     r.Driver,
		}.NewCommand("migrations"),
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Database) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Database) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
