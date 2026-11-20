package metrics

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

type cpuCollectorFunc func(context.Context) (telemetry.CPUUsage, error)

func (f cpuCollectorFunc) Collect(ctx context.Context) (telemetry.CPUUsage, error) {
	return f(ctx)
}

type memoryCollectorFunc func(context.Context) (telemetry.MemoryUsage, error)

func (f memoryCollectorFunc) Collect(ctx context.Context) (telemetry.MemoryUsage, error) {
	return f(ctx)
}

func TestRunFastStartsImmediatelyAndKeepsResultsIndependent(t *testing.T) {
	cpuErr := errors.New("CPU unavailable")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan FastSampleMsg, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunFast(
			ctx,
			time.Hour,
			cpuCollectorFunc(func(context.Context) (telemetry.CPUUsage, error) {
				return telemetry.CPUUsage{}, cpuErr
			}),
			memoryCollectorFunc(func(context.Context) (telemetry.MemoryUsage, error) {
				return telemetry.MemoryUsage{UsedPercent: 37.5}, nil
			}),
			func(sample FastSampleMsg) {
				result <- sample
				cancel()
			},
		)
	}()

	select {
	case sample := <-result:
		if !errors.Is(sample.CPU.Err, cpuErr) {
			t.Fatalf("CPU failure was not retained: %v", sample.CPU.Err)
		}
		if sample.Memory.Err != nil || sample.Memory.Usage.UsedPercent != 37.5 {
			t.Fatalf("memory result should succeed independently, got %+v", sample.Memory)
		}
		if sample.CPU.AttemptedAt.IsZero() || !sample.CPU.SampledAt.IsZero() {
			t.Fatalf("failed CPU sample should have an attempt time and no successful sample time: %+v", sample.CPU)
		}
		if sample.Memory.AttemptedAt.IsZero() || sample.Memory.SampledAt.IsZero() {
			t.Fatalf("memory sample should carry collection times: %+v", sample)
		}
	case <-time.After(time.Second):
		t.Fatal("first fast sample did not arrive immediately")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("fast worker did not stop after cancellation")
	}
}

func TestRunFastDoesNotOverlapAndExitsDuringCollection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{}, 4)
	workerErrors := make(chan string, 2)
	done := make(chan struct{})
	var calls atomic.Int32

	go func() {
		defer close(done)
		RunFast(
			ctx,
			time.Millisecond,
			cpuCollectorFunc(func(ctx context.Context) (telemetry.CPUUsage, error) {
				calls.Add(1)
				started <- struct{}{}
				<-ctx.Done()
				return telemetry.CPUUsage{}, ctx.Err()
			}),
			memoryCollectorFunc(func(context.Context) (telemetry.MemoryUsage, error) {
				workerErrors <- "memory collection started after cancellation during CPU"
				return telemetry.MemoryUsage{}, nil
			}),
			func(FastSampleMsg) { workerErrors <- "canceled collection published a sample" },
		)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("fast worker did not start its first collection immediately")
	}
	time.Sleep(20 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("a blocked collector was invoked %d times; want exactly one", got)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("fast worker did not exit after cancellation during collection")
	}
	select {
	case message := <-workerErrors:
		t.Fatal(message)
	default:
	}
}
