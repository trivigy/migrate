package types

import (
	"io"
)

type Executable interface {
	Execute(out io.Writer, args []string) error
}
