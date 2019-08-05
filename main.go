package migrate

import (
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/cluster"
	"github.com/trivigy/migrate/database"
	"github.com/trivigy/migrate/internal/nub"
)

// Registry is the container holding registered migrations.
var Registry sync.Map

// Migrate represents the migrate command object.
type Migrate struct {
	config map[string]Config
}

// NewMigrate instantiates a new migrate object and returns it.
func NewMigrate(config map[string]Config) Command {
	return &Migrate{config: config}
}

// NewCommand returns a new cobra.Command object.
func (r *Migrate) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Long:         "Devops and migrations management toolset",
		SilenceUsage: true,
	}

	clusterConfig := make(map[string]cluster.Config)
	for key, value := range r.config {
		clusterConfig[key] = value.Cluster
	}
	Cluster := cluster.NewCluster(clusterConfig).(*cluster.Cluster)

	databaseConfig := make(map[string]database.Config)
	for key, value := range r.config {
		databaseConfig[key] = value.Database
	}
	Database := database.NewDatabase(databaseConfig).(*database.Database)

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		Cluster.NewCommand("cluster"),
		Database.NewCommand("database"),
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
