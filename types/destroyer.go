package types

import (
	"io"
)

// Destroyer represents the destroying method interface for the driver.
type Destroyer interface {
	Destroy(out io.Writer) error
}
