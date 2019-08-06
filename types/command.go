package types

import (
	"io"
)

// Command represents an abstraction for a command.
type Command interface {
	Execute(name string, out io.Writer, args []string) error
}
