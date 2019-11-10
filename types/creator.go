package types

import (
	"context"
	"io"
)

// Creator represents the creation method interface for drivers.
type Creator interface {
	Create(ctx context.Context, out io.Writer) error
}
