package types

import (
	"io"
)

// Creator represents the creation method interface for drivers.
type Creator interface {
	Create(out io.Writer) error
}
