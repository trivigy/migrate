package resource

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/resource/kubernetes"
	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

// Kubernetes represents a kubernetes root command.
type Kubernetes struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Creator
		types.Destroyer
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Kubernetes)

// NewCommand returns a new cobra.Command object.
func (r Kubernetes) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " COMMAND",
		Short: "Kubernetes cluster release and deployment controller.",
		Long:  "Kubernetes cluster release and deployment controller",
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
		kubernetes.Releases{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("releases"),
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Kubernetes) Execute(name string, output io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(output)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Kubernetes) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
