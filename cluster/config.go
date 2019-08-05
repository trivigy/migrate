package cluster

import (
	"github.com/trivigy/migrate/driver"
)

// Config represents a configurations object of a cluster migration.
type Config struct {
	Migrations Migrations    `json:"migrations" yaml:"migrations"`
	Driver     driver.Driver `json:"driver" yaml:"driver"`
}
