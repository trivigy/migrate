package testutils

import (
	"context"
	"io"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/types"
)

type Database struct {
	Migrations *types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		driver.WithCreate
		driver.WithDestroy
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

func (r Database) Build() *databaseImpl {
	return &databaseImpl{
		migrations: r.Migrations,
		driver:     r.Driver,
	}
}

type databaseImpl struct {
	migrations *types.Migrations
	driver     interface {
		driver.WithCreate
		driver.WithDestroy
		driver.WithSource
	}
}

var _ interface {
	driver.WithCreate
	driver.WithDestroy
	driver.WithMigrations
	driver.WithSource
} = new(databaseImpl)

func (r databaseImpl) Migrations() *types.Migrations {
	return r.migrations
}

// Create executes the resource creation process.
func (r databaseImpl) Create(ctx context.Context, out io.Writer) error {
	return r.driver.Create(ctx, out)
}

// Destroy executes the resource destruction process.
func (r databaseImpl) Destroy(ctx context.Context, out io.Writer) error {
	return r.driver.Destroy(ctx, out)
}

// Source returns the data source name for the driver.
func (r databaseImpl) Source(ctx context.Context, out io.Writer) error {
	return r.driver.Source(ctx, out)
}
