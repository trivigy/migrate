// Package primitive implements basic set of commands.
package primitive

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Create represents the database create command object.
type Create struct {
	Driver interface {
		driver.WithCreate
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Create)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Create) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name[strings.LastIndex(name, ".")+1:],
		Short: "Constructs and starts a new instance of this resource.",
		Long:  "Constructs and starts a new instance of this resource",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			if patches, ok := r.Driver.(driver.WithPatches); ok {
				for _, patch := range *patches.Patches(name) {
					if err := patch.Do(ctx, cmd.OutOrStdout()); err != nil {
						return err
					}
				}
			}

			if try, _ := cmd.Flags().GetBool("try"); try {
				rbytes, err := yaml.Marshal(r.Driver)
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", string(rbytes))
				return nil
			}

			return r.run(ctx, cmd.OutOrStdout())
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool(
		"try", false,
		"Simulates and prints resource execution parameters.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Create) Execute(name string, out io.Writer, args []string) error {
	wrap := types.Executor{Name: name, Command: r}
	ctx := context.WithValue(context.Background(), global.RefRoot, wrap)
	cmd := r.NewCommand(ctx, name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Create) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the create command.
func (r Create) run(ctx context.Context, out io.Writer) error {
	if err := r.Driver.Create(ctx, out); err != nil {
		return err
	}
	return nil
}
