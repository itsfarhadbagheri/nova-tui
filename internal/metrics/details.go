package metrics

import (
	"context"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

const (
	ProcessesInterval   = 3 * time.Second
	TemperatureInterval = 5 * time.Second
)

type ProcessesSampleMsg = Sample[telemetry.Processes]
type TemperatureSampleMsg = Sample[[]telemetry.Temperature]

// RunSample polls a collector whose readings need no additional processing.
func RunSample[T any](ctx context.Context, interval time.Duration, collector Collector[T], publish func(Sample[T])) {
	if ctx == nil || collector == nil || publish == nil {
		return
	}
	poll(ctx, interval, func(ctx context.Context) Sample[T] {
		return collectSample(ctx, collector)
	}, publish)
}
