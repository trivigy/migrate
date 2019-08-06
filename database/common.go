package database

import (
	"fmt"
	"os"
	"sort"

	"github.com/trivigy/migrate/v2/internal/store"
	"github.com/trivigy/migrate/v2/internal/store/model"
	"github.com/trivigy/migrate/v2/types"
)

// common defines some shared methods used by database commands.
type common struct{}

// GenerateMigrationPlan creates a migration plan based on difference with the
// current state recorded on the database and direction.
func (r common) GenerateMigrationPlan(
	db *store.Context,
	direction types.Direction,
	migrations types.Migrations,
) ([]types.Migration, error) {
	if err := db.Migrations.CreateTableIfNotExists(); err != nil {
		return nil, err
	}

	sort.Sort(migrations)
	sortedRegistryMigrations := migrations
	sortedDatabaseMigrations, err := db.Migrations.GetMigrationsSorted()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	i := 0
	maxSize := max(len(sortedRegistryMigrations), len(sortedDatabaseMigrations))
	for ; i < maxSize; i++ {
		var rgMig *types.Migration
		if i < len(sortedRegistryMigrations) {
			rgMig = &sortedRegistryMigrations[i]
		}

		var dbMig *model.Migration
		if i < len(sortedDatabaseMigrations) {
			dbMig = &sortedDatabaseMigrations[i]
		}

		if rgMig != nil && dbMig != nil {
			if rgMig.Tag.String() != dbMig.Tag {
				return nil, fmt.Errorf(
					"migration tags mismatch %q != %q",
					rgMig.Tag.String(), dbMig.Tag,
				)
			}

		} else if rgMig != nil && dbMig == nil {
			if direction == types.DirectionUp {
				break
			} else {
				i--
				break
			}
		} else if rgMig == nil && dbMig != nil {
			return nil, fmt.Errorf("migration tags missing %q", dbMig.Tag)
		}
	}

	plan := make([]types.Migration, 0)
	if direction == types.DirectionUp {
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
	return plan, nil
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
