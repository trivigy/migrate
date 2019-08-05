package provider

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/trivigy/migrate/internal/retry"
)

// SQLDatabase represents an abstract remote sql database driver.
type SQLDatabase struct {
	Driver string
	Source string
}

// Setup executes the resource creation process.
func (r SQLDatabase) Setup(out io.Writer) error {
	db, err := sql.Open(r.Driver, r.Source)
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
func (r SQLDatabase) TearDown(out io.Writer) error {
	return nil
}
