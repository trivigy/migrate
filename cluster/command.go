package cluster

import (
	"github.com/trivigy/migrate/internal/nub"
)

// Command represents an abstraction for a command.
type Command interface {
	nub.Command
}
