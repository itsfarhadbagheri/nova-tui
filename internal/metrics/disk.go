package metrics

import (
	"context"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

type DiskSampleMsg = Sample[telemetry.DiskUsage]

// RunDisk polls a single-filesystem collector on its own cadence.
func RunDisk(ctx context.Context, interval time.Duration, collector Collector[telemetry.DiskUsage], publish func(DiskSampleMsg)) {
	RunSample(ctx, interval, collector, publish)
}
