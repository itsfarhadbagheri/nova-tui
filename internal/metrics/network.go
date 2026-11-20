package metrics

import (
	"context"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

type NetworkSampleMsg = Sample[telemetry.NetworkRates]

// RunNetwork converts successful interface counter snapshots to rates. Failed
// polls keep the previous baseline so the next rate spans the actual elapsed
// time since the last successful snapshot.
func RunNetwork(
	ctx context.Context,
	interval time.Duration,
	collector Collector[telemetry.NetworkCounters],
	publish func(NetworkSampleMsg),
) {
	if ctx == nil || collector == nil || publish == nil {
		return
	}

	var previous map[string]networkBaseline
	poll(ctx, interval, func(ctx context.Context) NetworkSampleMsg {
		counters := collectSample(ctx, collector)
		sample := NetworkSampleMsg{
			AttemptedAt: counters.AttemptedAt,
			SampledAt:   counters.SampledAt,
			Err:         counters.Err,
		}
		if counters.Err == nil {
			sample.Usage, previous = calculateNetworkRates(counters.Usage, previous, counters.SampledAt)
		}
		return sample
	}, publish)
}

type networkBaseline struct {
	bytesSent uint64
	bytesRecv uint64
	sampledAt time.Time
}

func calculateNetworkRates(
	counters telemetry.NetworkCounters,
	previous map[string]networkBaseline,
	sampledAt time.Time,
) (telemetry.NetworkRates, map[string]networkBaseline) {
	current := make(map[string]networkBaseline, len(counters.Interfaces))
	var rates telemetry.NetworkRates
	for _, iface := range counters.Interfaces {
		current[iface.Name] = networkBaseline{
			bytesSent: iface.BytesSent,
			bytesRecv: iface.BytesRecv,
			sampledAt: sampledAt,
		}

		prior, exists := previous[iface.Name]
		if !exists {
			continue
		}
		elapsed := sampledAt.Sub(prior.sampledAt)
		if received, ok := bytesPerSecond(prior.bytesRecv, iface.BytesRecv, elapsed); ok {
			rates.ReceivedBytesPerSecond += received
			rates.HasReceivedRate = true
		}
		if sent, ok := bytesPerSecond(prior.bytesSent, iface.BytesSent, elapsed); ok {
			rates.SentBytesPerSecond += sent
			rates.HasSentRate = true
		}
	}
	return rates, current
}

func bytesPerSecond(previous, current uint64, elapsed time.Duration) (float64, bool) {
	if elapsed <= 0 || current < previous {
		return 0, false
	}
	return float64(current-previous) / elapsed.Seconds(), true
}
