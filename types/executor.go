package types

import (
	"io"
)

// Executor is a wrapper class for passing a specialized executable root object
// across chained commands.
type Executor struct {
	Name    string
	Command Command
}

// Execute defines a method for executing the root command in the chainable
// command sequence.
func (r Executor) Execute(out io.Writer, args []string) error {
	return r.Command.Execute(r.Name, out, args)
}
