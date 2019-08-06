package config

import (
	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/types"
)

// Database defines database configuration to use for database migrations.
type Database struct {
	Migrations types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     driver.Database  `json:"driver" yaml:"driver"`
}
