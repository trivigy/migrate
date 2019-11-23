package types

import (
	"io"
)

// Executable defines an interface for the execute command of the root command
// executable method when it is passed through the chain of chained commands.
type Executable interface {
	Execute(out io.Writer, args []string) error
}
