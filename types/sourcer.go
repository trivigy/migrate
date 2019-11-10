package types

import (
	"context"
	"io"
)

// Sourcer represents a driver that is able to return its connection string.
type Sourcer interface {
	Source(ctx context.Context, out io.Writer) error
}
