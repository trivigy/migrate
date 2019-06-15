package store

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/gorp.v1"

	// mssql driver
	_ "github.com/denisenkom/go-mssqldb"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	// postgres driver
	_ "github.com/lib/pq"

	// sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

var supportedDialects = map[string]gorp.Dialect{
	"sqlite3":  gorp.SqliteDialect{},
	"postgres": gorp.PostgresDialect{},
	"mysql":    gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"},
	"mssql":    gorp.SqlServerDialect{},
}

// Context defines the global database context with access to all database
// available tables and operations.
type Context struct {
	db         *sql.DB
	dialect    gorp.Dialect
	Migrations Migrations
	Unittests  Unittests
}

// Open initializes the context and creates a database connection.
func Open(driver, source string) (*Context, error) {
	dialect := supportedDialects[driver]
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = db.Ping(); err != nil {
		return nil, errors.WithStack(err)
	}

	// When using the mysql driver, make sure that the parseTime option is
	// configured, otherwise it won't map time columns to time.Time. See
	// https://github.com/rubenv/sql-migrate/issues/2
	if _, ok := dialect.(gorp.MySQLDialect); ok {
		var out *time.Time
		if err := db.QueryRow("SELECT NOW()").Scan(&out); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	context := &Context{
		db:         db,
		dialect:    dialect,
		Migrations: Migrations{db, dialect},
		Unittests:  Unittests{db, dialect},
	}
	return context, nil
}

// GetDBMap returns the global connection database object.
func (r *Context) GetDBMap() *gorp.DbMap {
	return &gorp.DbMap{Db: r.db, Dialect: r.dialect}
}

// Close terminates the connection to the database and closes context.
func (r *Context) Close() error {
	return r.db.Close()
}
