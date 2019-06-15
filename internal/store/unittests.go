package store

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/gorp.v1"

	"github.com/trivigy/migrate/internal/dao"
)

const unittestsTableName = "unittests"

// Unittests defines a set of operations for the unittests table.
type Unittests struct {
	db      *sql.DB
	dialect gorp.Dialect
}

// GetDBMap returns the underlying unittests table database model object.
func (r Unittests) GetDBMap() *gorp.DbMap {
	dbMap := &gorp.DbMap{Db: r.db, Dialect: r.dialect}
	t := dbMap.AddTableWithName(dao.Unittest{}, unittestsTableName)
	t.SetKeys(false, "Value")
	return dbMap
}

// GetUnittests returns database unittest records.
func (r Unittests) GetUnittests() ([]dao.Unittest, error) {
	dbMap := r.GetDBMap()
	unittests := make([]dao.Unittest, 0)
	query := fmt.Sprintf(
		"SELECT * FROM %s",
		dbMap.Dialect.QuotedTableForQuery("", unittestsTableName),
	)
	if _, err := dbMap.Select(&unittests, query); err != nil {
		return nil, errors.WithStack(err)
	}
	return unittests, nil
}