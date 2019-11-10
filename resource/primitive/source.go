package primitive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
	"gopkg.in/go-playground/validator.v9"

	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Source represents the database source command object.
type Source struct {
	Driver interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Source)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Source) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Prints the data source name as a connection string.",
		Long:  "Prints the data source name as a connection string",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			merge, _ := cmd.Flags().GetStringSlice("merge")
			for _, set := range merge {
				split := strings.Split(set, "=")
				path := strings.TrimSpace(split[0])
				value := strings.TrimSpace(split[1])

				modifier, err := sjson.Set("", path, value)
				if err != nil {
					return err
				}

				if err := json.Unmarshal([]byte(modifier), r.Driver); err != nil {
					return err
				}
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				dump, err := json.Marshal(r.Driver)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", string(dump))
				return nil
			}

			validate := validator.New()
			if err := validate.Struct(r.Driver); err != nil {
				return fmt.Errorf("%w", err)
			}
			return r.run(context.Background(), cmd.OutOrStdout())
		},
		SilenceUsage: true,
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceP(
		"merge", "m", nil,
		"Merges specified json `PATH` with configured parameters.",
	)
	flags.Bool(
		"dry-run", false,
		"Simulate parameter merging without resource execution.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Source) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Source) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the source command.
func (r Source) run(ctx context.Context, out io.Writer) error {
	if err := r.Driver.Source(ctx, out); err != nil {
		return err
	}
	return nil
}
