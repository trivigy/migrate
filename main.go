package migrate

import (
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/cluster"
	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/database"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/types"
)

// Registry is the container holding registered migrations.
var Registry sync.Map

// Migrate represents the migrate command object.
type Migrate struct {
	config map[string]config.Migrate
}

// NewMigrate instantiates a new migrate object and returns it.
func NewMigrate(config map[string]config.Migrate) types.Command {
	return &Migrate{config: config}
}

// NewCommand returns a new cobra.Command object.
func (r *Migrate) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Version:      "2.0.0",
		Use:          name,
		Long:         "Kubernetes cluster releases and migrations administration toolset",
		SilenceUsage: true,
	}

	clusterConfig := make(map[string]config.Cluster)
	for key, value := range r.config {
		clusterConfig[key] = value.Cluster
	}

	databaseConfig := make(map[string]config.Database)
	for key, value := range r.config {
		databaseConfig[key] = value.Database
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		cluster.NewCluster(clusterConfig).(*cluster.Cluster).NewCommand("cluster"),
		database.NewDatabase(databaseConfig).(*database.Database).NewCommand("database"),
	)

	pflags := cmd.PersistentFlags()
	pflags.Bool("help", false, "Show help information.")
	pflags.StringP(
		"env", "e", nub.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolP(
		"version", "v", false,
		"Print version information and quit.",
	)
	return cmd
}

// Execute runs the command.
func (r *Migrate) Execute(name string, out io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(out)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
