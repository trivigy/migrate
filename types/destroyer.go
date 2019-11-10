package types

import (
	"context"
	"io"
)

// Destroyer represents the destroying method interface for the driver.
type Destroyer interface {
	Destroy(ctx context.Context, out io.Writer) error
}
