package store

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/gorp.v1"

	"github.com/trivigy/migrate/v2/internal/store/model"
)

const releasesTableName = "releases"

// Releases defines a wrapper struct for all of the releases table operations.
type Releases struct {
	db      *sql.DB
	dialect gorp.Dialect
}

// GetDBMap returns the underlying releases table database model object.
func (r Releases) GetDBMap() *gorp.DbMap {
	dbMap := &gorp.DbMap{Db: r.db, Dialect: r.dialect}
	t := dbMap.AddTableWithName(model.Release{}, releasesTableName)
	t.SetKeys(false, "Name", "Version")
	return dbMap
}

// CreateTableIfNotExists create releases table if one does not exist.
func (r Releases) CreateTableIfNotExists() error {
	dbMap := r.GetDBMap()
	if err := dbMap.CreateTablesIfNotExists(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DropTablesIfExists drops a table from the database if already exists.
func (r Releases) DropTablesIfExists() error {
	dbMap := r.GetDBMap()
	if err := dbMap.DropTablesIfExists(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Insert adds a release record to the database.
func (r Releases) Insert(releases ...interface{}) error {
	dbMap := r.GetDBMap()
	if err := dbMap.Insert(releases...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Delete instructs a release record to be deleted from the database.
func (r Releases) Delete(releases ...interface{}) error {
	dbMap := r.GetDBMap()
	if _, err := dbMap.Delete(releases...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetReleases returns database release records.
func (r Releases) GetReleases() ([]model.Release, error) {
	dbMap := r.GetDBMap()
	releases := make([]model.Release, 0)
	query := fmt.Sprintf(
		`SELECT * FROM %s`,
		dbMap.Dialect.QuotedTableForQuery("", migrationsTableName),
	)
	if _, err := dbMap.Select(&releases, query); err != nil {
		return nil, errors.WithStack(err)
	}
	return releases, nil
}

// GetMigrationsSorted returns sorted database release records.
func (r Releases) GetMigrationsSorted() (model.Releases, error) {
	clusterReleases, err := r.GetReleases()
	if err != nil {
		return nil, err
	}

	sortedClusterReleases := model.Releases(clusterReleases)
	sort.Sort(sortedClusterReleases)
	return sortedClusterReleases, nil
}
