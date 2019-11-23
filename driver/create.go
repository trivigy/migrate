// Package driver implements driver capability interfaces and drivers for local development.
package driver

import (
	"context"
	"io"
)

// WithCreate represents the creation method interface for a driver.
type WithCreate interface {
	Create(ctx context.Context, out io.Writer) error
}
