// Package store implements operations associated with update migration records.
package store

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"time"

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
	Releases   Releases
	Unittests  Unittests
}

// Open initializes the context and creates a database connection.
func Open(driver, source string) (*Context, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dialect := supportedDialects[getDriverName(db.Driver())]

	// When using the mysql driver, make sure that the parseTime option is
	// configured, otherwise it won't map time columns to time.Time. See
	// https://github.com/rubenv/sql-migrate/issues/2
	if _, ok := dialect.(gorp.MySQLDialect); ok {
		var out *time.Time
		if err := db.QueryRow("SELECT NOW()").Scan(&out); err != nil {
			return nil, err
		}
	}

	context := &Context{
		db:         db,
		dialect:    dialect,
		Migrations: Migrations{db, dialect},
		Releases:   Releases{db, dialect},
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

func getDriverName(driver driver.Driver) string {
	registeredDriverNamesByType := map[reflect.Type]string{}
	for _, driverName := range sql.Drivers() {
		if db, _ := sql.Open(driverName, ""); db != nil {
			driverType := reflect.TypeOf(db.Driver())
			registeredDriverNamesByType[driverType] = driverName
		}
	}
	driverType := reflect.TypeOf(driver)
	if driverName, found := registeredDriverNamesByType[driverType]; found {
		return driverName
	}
	return ""
}
