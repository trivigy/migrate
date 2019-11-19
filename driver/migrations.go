package driver

import (
	"github.com/trivigy/migrate/v2/types"
)

type WithMigrations interface {
	Migrations() *types.Migrations
}
