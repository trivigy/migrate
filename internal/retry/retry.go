package retry

import (
	"context"
	"time"
)

// Do executes a function at a given interval. The context may be used to cancel
// the execution. For example for timing out a long running repeating task.
func Do(ctx context.Context, interval time.Duration, fn func() (bool, error)) error {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if cond, err := fn(); !cond && err == nil {
				ticker.Stop()
				return nil
			}
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}
}
