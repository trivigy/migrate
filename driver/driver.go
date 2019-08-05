package driver

import (
	"io"
)

// Driver represents an interface to an abstruct automation driver.
type Driver interface {
	Setup(out io.Writer) error
	TearDown(out io.Writer) error
}
