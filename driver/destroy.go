package driver

import (
	"context"
	"io"
)

// WithDestroy represents the destroying method interface for a driver.
type WithDestroy interface {
	Destroy(ctx context.Context, out io.Writer) error
}
