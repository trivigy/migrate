package provider

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/trivigy/migrate/v2/internal/retry"
)

// SQL represents an abstract remote sql database driver.
type SQL struct {
	Dialect    string `json:"dialect" yaml:"dialect"`
	DataSource string `json:"dataSource" yaml:"dataSource"`
}

// Setup executes the resource creation process.
func (r SQL) Setup(out io.Writer) error {
	db, err := sql.Open(r.Dialect, r.DataSource)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := retry.Do(ctx, 2*time.Second, func() (bool, error) {
		err := db.Ping()
		if err == nil {
			return false, nil
		}
		return true, err
	}); err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		return err
	}
	return nil
}

// TearDown executes the resource destruction process.
func (r SQL) TearDown(out io.Writer) error {
	return nil
}

// Name returns the driver name.
func (r SQL) Name() string {
	return r.Dialect

}

// Source returns the data source name for the driver.
func (r SQL) Source() (string, error) {
	return r.DataSource, nil
}
