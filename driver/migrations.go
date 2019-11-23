package driver

import (
	"github.com/trivigy/migrate/v2/types"
)

// WithMigrations represents the method interface for extracting the migrations
// from a driver. This method is likely to be used by a database driver which
// manipulates migrations.
type WithMigrations interface {
	Migrations() *types.Migrations
}
