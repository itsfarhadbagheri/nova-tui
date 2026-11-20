package metrics

import (
	"context"
	"time"
)

type Collector[T any] interface {
	Collect(context.Context) (T, error)
}

// Sample records a collection attempt. SampledAt is set only on success.
type Sample[T any] struct {
	Usage       T
	SampledAt   time.Time
	AttemptedAt time.Time
	Err         error
}

// HasAttempt reports whether this sample represents a collection attempt.
func (s Sample[T]) HasAttempt() bool {
	return s.Err != nil || !s.AttemptedAt.IsZero() || !s.SampledAt.IsZero()
}

func collectSample[T any](ctx context.Context, collector Collector[T]) Sample[T] {
	sample := Sample[T]{AttemptedAt: time.Now()}
	usage, err := collector.Collect(ctx)
	sample.Err = err
	if err == nil {
		sample.Usage = usage
		sample.SampledAt = time.Now()
	}
	return sample
}
