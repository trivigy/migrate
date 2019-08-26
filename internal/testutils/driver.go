package testutils

import (
	"io"
)

// Driver defines a placeholder testing driver.
type Driver struct{}

// Create is a placeholder for the Creator interface.
func (r Driver) Create(out io.Writer) error {
	panic("implement me")
}

// Destroy is a placeholder for the Destroyer interface.
func (r Driver) Destroy(out io.Writer) error {
	panic("implement me")
}

// Source is a placeholder for the Sourced interface.
func (r Driver) Source() (string, error) {
	panic("implement me")
}
