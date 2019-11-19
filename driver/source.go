package driver

import (
	"context"
	"io"
)

// WithSource represents a driver that is able to return its connection string.
type WithSource interface {
	Source(ctx context.Context, out io.Writer) error
}
