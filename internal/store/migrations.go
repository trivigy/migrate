package store

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/gorp.v1"

	"github.com/trivigy/migrate/internal/dao"
)

const migrationsTableName = "__migrations"

// Migrations defines a wrapper struct for all of the migrations table
// operations.
type Migrations struct {
	db      *sql.DB
	dialect gorp.Dialect
}

// GetDBMap returns the underlying migrations table database model object.
func (r Migrations) GetDBMap() *gorp.DbMap {
	dbMap := &gorp.DbMap{Db: r.db, Dialect: r.dialect}
	t := dbMap.AddTableWithName(dao.Migration{}, migrationsTableName)
	t.SetKeys(false, "Tag")
	return dbMap
}

// CreateTableIfNotExists create migrations table if one does not exist.
func (r Migrations) CreateTableIfNotExists() error {
	dbMap := r.GetDBMap()
	if err := dbMap.CreateTablesIfNotExists(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DropTablesIfExists drops a table from the database if already exists.
func (r Migrations) DropTablesIfExists() error {
	dbMap := r.GetDBMap()
	if err := dbMap.DropTablesIfExists(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Insert adds a migration record to the database.
func (r Migrations) Insert(migrations ...interface{}) error {
	dbMap := r.GetDBMap()
	if err := dbMap.Insert(migrations...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Delete instructs a migration record to be deleted from the database.
func (r Migrations) Delete(migrations ...interface{}) error {
	dbMap := r.GetDBMap()
	if _, err := dbMap.Delete(migrations...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetMigrations returns database migration records.
func (r Migrations) GetMigrations() ([]dao.Migration, error) {
	dbMap := r.GetDBMap()
	migrations := make([]dao.Migration, 0)
	query := fmt.Sprintf(
		`SELECT * FROM %s`,
		dbMap.Dialect.QuotedTableForQuery("", migrationsTableName),
	)
	if _, err := dbMap.Select(&migrations, query); err != nil {
		return nil, errors.WithStack(err)
	}
	return migrations, nil
}

// GetMigrationsSorted returns sorted database migration records.
func (r Migrations) GetMigrationsSorted() (dao.Migrations, error) {
	databaseMigrations, err := r.GetMigrations()
	if err != nil {
		return nil, err
	}

	sortedDatabaseMigrations := dao.Migrations(databaseMigrations)
	sort.Sort(sortedDatabaseMigrations)
	return sortedDatabaseMigrations, nil
}
