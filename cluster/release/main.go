package release

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/types"
)

// Release represents a cluster release root command.
type Release struct {
	config map[string]config.Cluster
}

// NewRelease instantiates a new release command and returns it.
func NewRelease(config map[string]config.Cluster) types.Command {
	return &Release{config: config}
}

// NewCommand returns a new cobra.Command object.
func (r *Release) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "Clusters migration tool.",
		Long:         "Clusters migration tool",
		SilenceUsage: true,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		NewGenerate(r.config).(*Generate).NewCommand("generate"),
		NewInstall(r.config).(*Install).NewCommand("install"),
		NewUpgrade(r.config).(*Upgrade).NewCommand("upgrade"),
		NewDelete(r.config).(*Delete).NewCommand("delete"),
		NewList(r.config).(*List).NewCommand("list"),
		NewInspect(r.config).(*Inspect).NewCommand("inspect"),
		NewHistory(r.config).(*History).NewCommand("history"),
	)

	pflags := cmd.PersistentFlags()
	pflags.Bool("help", false, "Show help information.")
	pflags.StringP(
		"env", "e", nub.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	return cmd
}

// Execute runs the command.
func (r *Release) Execute(name string, output io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(output)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
