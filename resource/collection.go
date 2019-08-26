package resource

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/types"
)

// Collection represents a collection aggregator command.
type Collection map[string]types.Resource

// NewCommand returns a new cobra.Command object.
func (r Collection) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use: name,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	for name, resource := range r {
		cmd.AddCommand(resource.NewCommand(name))
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Collection) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}
