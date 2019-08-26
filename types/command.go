package types

import (
	"io"
)

// Command defines an interface for a command which is executable.
type Command interface {
	Execute(name string, out io.Writer, args []string) error
}
