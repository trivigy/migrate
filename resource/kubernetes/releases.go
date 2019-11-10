// Package kubernetes implements kubernetes related set of commands.
package kubernetes

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/resource/kubernetes/releases"
	"github.com/trivigy/migrate/v2/types"
)

// Releases represents a cluster release root command.
type Releases struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Releases)

// NewCommand returns a new cobra.Command object.
func (r Releases) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " COMMAND",
		Short: "Manages the lifecycle of a kubernetes release.",
		Long:  "Manages the lifecycle of a kubernetes release",
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
		releases.Generate{
			Releases: r.Releases,
		}.NewCommand("generate"),
		releases.Install{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("install"),
		releases.Upgrade{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("upgrade"),
		releases.Uninstall{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("uninstall"),
		releases.List{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("list"),
		releases.Describe{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("describe"),
		releases.History{
			Namespace: r.Namespace,
			Releases:  r.Releases,
			Driver:    r.Driver,
		}.NewCommand("history"),
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Releases) Execute(name string, output io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(output)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Releases) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}
