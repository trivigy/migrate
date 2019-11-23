package types

import (
	"context"
	"io"
	"sync"
)

// Patch defines an executable patch function and filters for which this
// function will run.
type Patch struct {
	once    sync.Once
	Filters []string
	Func    func(ctx context.Context, out io.Writer) error
}

// Do executes the patch function once and prevents repeated execution.
func (r *Patch) Do(ctx context.Context, out io.Writer) error {
	var err error
	r.once.Do(func() {
		err = r.Func(ctx, out)
	})
	if err != nil {
		return err
	}
	return nil
}
