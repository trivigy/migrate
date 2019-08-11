package config

import (
	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/types"
)

// Cluster represents a configurations object of a cluster migration.
type Cluster struct {
	Namespace string          `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    driver.Cluster  `json:"driver" yaml:"driver"`
}
