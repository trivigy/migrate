package kubernetes

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/resource/kubernetes/releases"
	"github.com/trivigy/migrate/v2/types"
)

// Releases represents a cluster release root command.
type Releases struct {
	Namespace string         `json:"namespace" yaml:"namespace"`
	Releases  types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.KubeConfiged
	} `json:"driver" yaml:"driver"`
}

// NewCommand returns a new cobra.Command object.
func (r Releases) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "Manages the lifecycle of a kubernetes release.",
		Long:         "Manages the lifecycle of a kubernetes release",
		SilenceUsage: true,
	}

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
