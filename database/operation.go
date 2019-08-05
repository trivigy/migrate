package database

import (
	"database/sql"
	"fmt"

	"gopkg.in/gorp.v1"

	"github.com/trivigy/migrate/internal/enum"
	"github.com/trivigy/migrate/internal/store"
)

// Operation defines a single query operation to run on the database.
type Operation struct {
	Query     string `json:"query" yaml:"query"`
	DisableTx bool   `json:"disableTx" yaml:"disableTx"`
}

// Execute runs the query operation on the database.
func (r Operation) Execute(db *store.Context, migration Migration, d enum.Direction) error {
	var err error
	var executor Executor

	dbMap := db.GetDBMap()
	if r.DisableTx {
		executor = dbMap
	} else {
		if executor, err = dbMap.Begin(); err != nil {
			return fmt.Errorf(
				"transaction begin failed %q (%s)",
				migration.Tag.String()+"_"+migration.Name, d,
			)
		}
	}

	if _, err := executor.Exec(r.Query); err != nil {
		if tx, ok := executor.(*gorp.Transaction); ok {
			if err = tx.Rollback(); err != nil {
				return fmt.Errorf(
					"transaction rollback failed %q (%s)",
					migration.Tag.String()+"_"+migration.Name, d,
				)
			}
		}
		return fmt.Errorf("migration query failed %q (%s)\n%s",
			migration.Tag.String()+"_"+migration.Name, d, r.Query,
		)
	}

	if tx, ok := executor.(*gorp.Transaction); ok {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("transaction commit failed %q (%s)",
				migration.Tag.String()+"_"+migration.Name, d,
			)
		}
	}
	return nil
}

// Executor describes an abstract database operations executor.
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Insert(list ...interface{}) error
	Delete(list ...interface{}) (int64, error)
}
