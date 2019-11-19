package types

import (
	"io"
)

type Executor struct {
	Name    string
	Command Command
}

func (r Executor) Execute(out io.Writer, args []string) error {
	return r.Command.Execute(r.Name, out, args)
}
