package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/trivigy/migrate/internal/dao"
	"github.com/trivigy/migrate/internal/dto"
	"github.com/trivigy/migrate/internal/enum"
	"github.com/trivigy/migrate/internal/store"
)

var (
	db       *store.Context
	registry map[string]dto.Migration
	configs  *viper.Viper
	root     *rootCommand
	create   *createCommand
	down     *downCommand
	status   *statusCommand
	up       *upCommand
)

func init() {
	ReInitialize()
}

// ReInitialize restarts the application with all the command flags,
// configurations, and migration registory reset. This is primarily useful for
// testing.
func ReInitialize() {
	registry = map[string]dto.Migration{}
	configs = viper.New()
	create = newCreateCommand()
	down = newDownCommand()
	status = newStatusCommand()
	up = newUpCommand()
	root = newRootCommand()
}

// SetConfigs allows for passing database connection configurations with custom
// environment names.
func SetConfigs(rbytes []byte) error {
	configs.SetConfigType("yaml")
	if err := configs.ReadConfig(bytes.NewBuffer(rbytes)); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Execute runs the main application loop.
func Execute() error {
	if err := root.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Append allows for adding migrations to the migration registry list.
func Append(migration dto.Migration) error {
	if _, loaded := registry[migration.Tag.String()]; loaded {
		return errors.Errorf("duplicate migration tag %q", migration.Tag)
	}
	registry[migration.Tag.String()] = migration
	return nil
}

func generateMigrationPlan(cmd *cobra.Command, db *store.Context, direction enum.Direction) []dto.Migration {
	if err := db.Migrations.CreateTableIfNotExists(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	sortedRegistryMigrations := getSortedRegistryMigrations()
	sortedDatabaseMigrations, err := db.Migrations.GetMigrationsSorted()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	i := 0
	maxSize := max(len(sortedRegistryMigrations), len(sortedDatabaseMigrations))
	for ; i < maxSize; i++ {
		var rgMig *dto.Migration
		if i < len(sortedRegistryMigrations) {
			rgMig = &sortedRegistryMigrations[i]
		}

		var dbMig *dao.Migration
		if i < len(sortedDatabaseMigrations) {
			dbMig = &sortedDatabaseMigrations[i]
		}

		if rgMig != nil && dbMig != nil {
			if rgMig.Tag.String() != dbMig.Tag {
				fmt.Fprintf(cmd.OutOrStderr(),
					"error: migration tags mismatch %q != %q\n",
					rgMig.Tag.String(), dbMig.Tag,
				)
				os.Exit(1)
			}

		} else if rgMig != nil && dbMig == nil {
			if direction == enum.DirectionUp {
				break
			} else {
				i--
				break
			}
		} else if rgMig == nil && dbMig != nil {
			fmt.Fprintf(cmd.OutOrStderr(),
				"error: migration tags missing %q\n", dbMig.Tag,
			)
			os.Exit(1)
		}
	}

	plan := make([]dto.Migration, 0)
	if direction == enum.DirectionUp {
		for j := i; j < len(sortedRegistryMigrations); j++ {
			plan = append(plan, sortedRegistryMigrations[j])
		}
	} else {
		if i > len(sortedDatabaseMigrations)-1 {
			i = len(sortedDatabaseMigrations) - 1
		}

		for j := i; j >= 0; j-- {
			plan = append(plan, sortedRegistryMigrations[j])
		}
	}
	return plan
}

func getSortedRegistryMigrations() dto.Migrations {
	migrations := make(dto.Migrations, 0, len(registry))
	for _, migration := range registry {
		migrations = append(migrations, migration)
	}
	sort.Sort(migrations)
	return migrations
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
