package types

import (
	"context"
	"io"
	"sync"
)

type Patch struct {
	once    sync.Once
	Filters []string
	Func    func(ctx context.Context, out io.Writer) error
}

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
