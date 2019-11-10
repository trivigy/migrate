package resource

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

// DomainName represents a database root command.
type DomainName struct {
	Driver interface {
		types.Creator
		types.Destroyer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(DomainName)

// NewCommand returns a new cobra.Command object.
func (r DomainName) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " COMMAND",
		Short: "Controls instance of domain name service resource.",
		Long:  "Controls instance of domain name service resource",
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
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r DomainName) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r DomainName) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
