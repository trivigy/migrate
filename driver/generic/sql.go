// Package generic implements abstract drivers for migrate package.
package generic

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/trivigy/migrate/v2/internal/retry"
	"github.com/trivigy/migrate/v2/types"
)

// SQL represents an abstract remote sql database driver.
type SQL struct {
	Dialect    string `json:"dialect" yaml:"dialect"`
	DataSource string `json:"dataSource" yaml:"dataSource"`
}

var _ interface {
	types.Creator
	types.Destroyer
	types.Sourcer
} = new(SQL)

// Create executes the resource creation process.
func (r SQL) Create(ctx context.Context, out io.Writer) error {
	db, err := sql.Open(r.Dialect, r.DataSource)
	if err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
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

// Destroy executes the resource destruction process.
func (r SQL) Destroy(ctx context.Context, out io.Writer) error {
	return nil
}

// // Name returns the driver name.
// func (r SQL) Name() string {
// 	return r.Dialect
// }

// Source returns the data source name for the driver.
func (r SQL) Source(ctx context.Context, out io.Writer) error {
	if _, err := out.Write([]byte(r.DataSource)); err != nil {
		return err
	}
	return nil
}
