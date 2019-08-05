package migrate

import (
	"github.com/trivigy/migrate/cluster"
	"github.com/trivigy/migrate/database"
)

// Config represents a configurations object of a migrate command.
type Config struct {
	Cluster  cluster.Config  `json:"cluster" yaml:"cluster"`
	Database database.Config `json:"database" yaml:"database"`
}
