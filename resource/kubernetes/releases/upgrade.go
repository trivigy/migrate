package releases

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Upgrade represents the cluster release upgrade command object.
type Upgrade struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

// UpgradeOptions is used for executing the run() command.
type UpgradeOptions struct{}

var _ interface {
	types.Resource
	types.Command
} = new(Upgrade)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Upgrade) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Redeploy a modified release and track revision version.",
		Long:  "Redeploy a modified release and track revision version",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := UpgradeOptions{}
			return r.run(cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Upgrade) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Upgrade) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the create command.
func (r Upgrade) run(out io.Writer, opts UpgradeOptions) error {
	return nil
}
