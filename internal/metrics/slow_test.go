package metrics

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

func TestCalculateNetworkRatesHandlesWarmupResetAndInterfaceRemoval(t *testing.T) {
	start := time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)
	previous := map[string]networkBaseline{}
	first, previous := calculateNetworkRates(telemetry.NetworkCounters{Interfaces: []telemetry.InterfaceCounters{
		{Name: "eth0", BytesRecv: 100, BytesSent: 200},
		{Name: "wlan0", BytesRecv: 300, BytesSent: 400},
	}}, previous, start)
	if first.HasReceivedRate || first.HasSentRate {
		t.Fatalf("first counter sample should only establish a baseline, got %+v", first)
	}

	second, previous := calculateNetworkRates(telemetry.NetworkCounters{Interfaces: []telemetry.InterfaceCounters{
		{Name: "eth0", BytesRecv: 100, BytesSent: 600},
		{Name: "wlan0", BytesRecv: 100, BytesSent: 400},
	}}, previous, start.Add(2*time.Second))
	if !second.HasReceivedRate || second.ReceivedBytesPerSecond != 0 {
		t.Fatalf("zero-byte deltas are valid rates: %+v", second)
	}
	if !second.HasSentRate || second.SentBytesPerSecond != 200 {
		t.Fatalf("send rate should sum independently across interfaces, got %+v", second)
	}

	third, previous := calculateNetworkRates(telemetry.NetworkCounters{Interfaces: []telemetry.InterfaceCounters{
		{Name: "eth0", BytesRecv: 200, BytesSent: 500},
	}}, previous, start.Add(3*time.Second))
	if !third.HasReceivedRate || third.ReceivedBytesPerSecond != 100 {
		t.Fatalf("receive rate should remain valid when the send counter resets: %+v", third)
	}
	if third.HasSentRate {
		t.Fatalf("a reset send counter must warm up again: %+v", third)
	}
	if len(previous) != 1 {
		t.Fatalf("removed interfaces should leave the replacement baseline, got %d entries", len(previous))
	}

	fourth, previous := calculateNetworkRates(telemetry.NetworkCounters{Interfaces: []telemetry.InterfaceCounters{
		{Name: "eth0", BytesRecv: 300, BytesSent: 600},
		{Name: "wlan0", BytesRecv: 500, BytesSent: 1000},
	}}, previous, start.Add(4*time.Second))
	if !fourth.HasReceivedRate || fourth.ReceivedBytesPerSecond != 100 {
		t.Fatalf("returned interface should warm up while existing interface stays comparable: %+v", fourth)
	}
	if !fourth.HasSentRate || fourth.SentBytesPerSecond != 100 {
		t.Fatalf("new interface should not replace valid rates from an existing interface: %+v", fourth)
	}
	if len(previous) != 2 {
		t.Fatalf("returned interface should enter the new baseline, got %d entries", len(previous))
	}
}

func TestBytesPerSecondRejectsResetAndZeroElapsed(t *testing.T) {
	if _, ok := bytesPerSecond(10, 20, 0); ok {
		t.Fatal("zero elapsed time must not produce a rate")
	}
	if _, ok := bytesPerSecond(20, 10, time.Second); ok {
		t.Fatal("a reset counter must not produce a rate")
	}
	rate, ok := bytesPerSecond(10, 10, time.Second)
	if !ok || rate != 0 {
		t.Fatalf("unchanged counters should produce a valid zero rate, got %v, %v", rate, ok)
	}
}

func TestRunNetworkRetainsBaselineAcrossFailedPoll(t *testing.T) {
	collector := &sequenceNetworkCollector{results: []networkResult{
		{counters: telemetry.NetworkCounters{Interfaces: []telemetry.InterfaceCounters{{Name: "eth0", BytesRecv: 1000, BytesSent: 2000}}}},
		{err: errors.New("temporary counter read failure")},
		{counters: telemetry.NetworkCounters{Interfaces: []telemetry.InterfaceCounters{{Name: "eth0", BytesRecv: 3000, BytesSent: 6000}}}},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results := make(chan NetworkSampleMsg, 3)
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunNetwork(ctx, 10*time.Millisecond, collector, func(sample NetworkSampleMsg) { results <- sample })
	}()

	first := receiveNetworkSample(t, results)
	failed := receiveNetworkSample(t, results)
	third := receiveNetworkSample(t, results)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("network worker did not stop after cancellation")
	}

	if first.Err != nil || first.Usage.HasReceivedRate || first.Usage.HasSentRate {
		t.Fatalf("first poll should warm up the baseline, got %+v", first)
	}
	if failed.Err == nil {
		t.Fatal("middle poll should report the injected error")
	}
	if third.Err != nil || !third.Usage.HasReceivedRate || !third.Usage.HasSentRate {
		t.Fatalf("successful poll after failure should recover using the retained baseline: %+v", third)
	}
	elapsed := third.SampledAt.Sub(first.SampledAt).Seconds()
	expectedReceive := 2000 / elapsed
	expectedSend := 4000 / elapsed
	if math.Abs(third.Usage.ReceivedBytesPerSecond-expectedReceive) > 0.01 ||
		math.Abs(third.Usage.SentBytesPerSecond-expectedSend) > 0.01 {
		t.Fatalf("rates should span successful snapshots across the failed poll; got %+v, elapsed %.3fs", third.Usage, elapsed)
	}
}

func TestSlowDiskPollDoesNotBlockFastWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	diskStarted := make(chan struct{}, 1)
	diskDone := make(chan struct{})
	go func() {
		defer close(diskDone)
		RunDisk(ctx, time.Second, diskCollectorFunc(func(ctx context.Context) (telemetry.DiskUsage, error) {
			diskStarted <- struct{}{}
			<-ctx.Done()
			return telemetry.DiskUsage{}, ctx.Err()
		}), func(DiskSampleMsg) {})
	}()
	select {
	case <-diskStarted:
	case <-time.After(time.Second):
		t.Fatal("disk worker did not start immediately")
	}

	fastSample := make(chan FastSampleMsg, 1)
	fastDone := make(chan struct{})
	go func() {
		defer close(fastDone)
		RunFast(
			ctx,
			time.Second,
			cpuCollectorFunc(func(context.Context) (telemetry.CPUUsage, error) {
				return telemetry.CPUUsage{Usage: 5}, nil
			}),
			memoryCollectorFunc(func(context.Context) (telemetry.MemoryUsage, error) {
				return telemetry.MemoryUsage{UsedPercent: 50}, nil
			}),
			func(sample FastSampleMsg) { fastSample <- sample },
		)
	}()
	select {
	case sample := <-fastSample:
		if sample.CPU.Err != nil || sample.Memory.Err != nil {
			t.Fatalf("fast sample should complete while disk is blocked: %+v", sample)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked disk collection delayed CPU and memory sampling")
	}

	cancel()
	for _, done := range []<-chan struct{}{diskDone, fastDone} {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("worker did not stop after cancellation")
		}
	}
}

type networkResult struct {
	counters telemetry.NetworkCounters
	err      error
}

type sequenceNetworkCollector struct {
	mu      sync.Mutex
	results []networkResult
}

func (c *sequenceNetworkCollector) Collect(context.Context) (telemetry.NetworkCounters, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.results) == 0 {
		return telemetry.NetworkCounters{}, errors.New("unexpected extra network collection")
	}
	result := c.results[0]
	c.results = c.results[1:]
	return result.counters, result.err
}

type diskCollectorFunc func(context.Context) (telemetry.DiskUsage, error)

func (f diskCollectorFunc) Collect(ctx context.Context) (telemetry.DiskUsage, error) {
	return f(ctx)
}

func receiveNetworkSample(t *testing.T, samples <-chan NetworkSampleMsg) NetworkSampleMsg {
	t.Helper()
	select {
	case sample := <-samples:
		return sample
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for network sample")
		return NetworkSampleMsg{}
	}
}
