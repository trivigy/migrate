// Package migrate implements an abstract migration tool.
package migrate

import (
	"github.com/trivigy/migrate/v2/types"
)

// Registry is the container holding registered migrations.
var Registry types.Registry
