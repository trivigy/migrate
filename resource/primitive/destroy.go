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

// Destroy represents the database destroy command.
type Destroy struct {
	Driver interface {
		types.Destroyer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Destroy)

// NewCommand returns a new cobra.Command destroy command object.
func (r Destroy) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Stops and removes running instance of this resource.",
		Long:  "Stops and removes running instance of this resource",
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
func (r Destroy) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Destroy) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the destroy command.
func (r Destroy) run(ctx context.Context, out io.Writer) error {
	if err := r.Driver.Destroy(ctx, out); err != nil {
		return err
	}
	return nil
}
