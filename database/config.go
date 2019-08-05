package database

import (
	"github.com/trivigy/migrate/driver"
)

// Config defines database configuration to use for database migrations.
type Config struct {
	Migrations Migrations      `json:"migrations" yaml:"migrations"`
	Driver     driver.Database `json:"driver" yaml:"driver"`
}
