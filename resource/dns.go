package resource

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

// DNS represents a database root command.
type DNS struct {
	Driver interface {
		types.Creator
		types.Destroyer
	} `json:"driver" yaml:"driver"`
}

// NewCommand returns a new cobra.Command object.
func (r DNS) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "Controls instance of domain name system resource.",
		Long:         "Controls instance of domain name system resource",
		SilenceUsage: true,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		primitive.Create{
			Driver: r.Driver,
		}.NewCommand("create"),
		primitive.Destroy{
			Driver: r.Driver,
		}.NewCommand("destroy"),
	)

	pflags := cmd.Flags()
	pflags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r DNS) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return err
	}
	return nil
}
