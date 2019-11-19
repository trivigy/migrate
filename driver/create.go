package driver

import (
	"context"
	"io"
)

// WithCreate represents the creation method interface for drivers.
type WithCreate interface {
	Create(ctx context.Context, out io.Writer) error
}
