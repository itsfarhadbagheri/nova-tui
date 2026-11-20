package metrics

import (
	"context"
	"time"
)

const (
	FastInterval    = time.Second
	DiskInterval    = 10 * time.Second
	NetworkInterval = 2 * time.Second
)

// poll collects immediately and waits interval after each publication.
func poll[T any](ctx context.Context, interval time.Duration, collect func(context.Context) T, publish func(T)) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		if ctx.Err() != nil {
			return
		}

		sample := collect(ctx)
		if ctx.Err() != nil {
			return
		}
		publish(sample)
		timer.Reset(interval)
	}
}
