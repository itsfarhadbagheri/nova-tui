package metrics

import (
	"context"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

type CPUUsageSample = Sample[telemetry.CPUUsage]
type MemoryUsageSample = Sample[telemetry.MemoryUsage]

// FastSampleMsg is one completed fast-poll cycle. The CPU and memory results
// remain independent so a failure in one collector does not hide the other.
type FastSampleMsg struct {
	CPU    CPUUsageSample
	Memory MemoryUsageSample
}

// RunFast collects CPU and memory in sequence and publishes each completed
// cycle before waiting interval. The first cycle starts immediately.
func RunFast(
	ctx context.Context,
	interval time.Duration,
	cpu Collector[telemetry.CPUUsage],
	memory Collector[telemetry.MemoryUsage],
	publish func(FastSampleMsg),
) {
	if ctx == nil || cpu == nil || memory == nil || publish == nil {
		return
	}

	poll(ctx, interval, func(ctx context.Context) FastSampleMsg {
		sample := FastSampleMsg{CPU: collectSample(ctx, cpu)}
		if ctx.Err() == nil {
			sample.Memory = collectSample(ctx, memory)
		}
		return sample
	}, publish)
}
