package uistarter

import (
	"errors"
	"math"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/metrics"
	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui/pages/dashboard"
)

type reading[T any] struct {
	value     T
	sampledAt time.Time
	attempted bool
	valid     bool
	failed    bool
}

func (r *reading[T]) apply(value T, sampledAt, attemptedAt, now time.Time, err error) {
	r.attempted = true
	if err != nil {
		r.failed = true
		return
	}

	if sampledAt.IsZero() {
		sampledAt = attemptedAt
		if sampledAt.IsZero() {
			sampledAt = now
		}
	}
	r.value = value
	r.sampledAt = sampledAt
	r.valid = true
	r.failed = false
}

func (r reading[T]) status(now time.Time, interval time.Duration) ui.ReadingStatus {
	if !r.valid {
		if !r.attempted {
			return ui.ReadingStatusWaiting
		}
		return ui.ReadingStatusUnavailable
	}
	if r.failed || (interval > 0 && now.Sub(r.sampledAt) > 2*interval) {
		return ui.ReadingStatusStale
	}
	return ui.ReadingStatusCurrent
}

func (r reading[T]) view(now time.Time, interval time.Duration) ui.Reading[T] {
	return ui.Reading[T]{Value: r.value, Status: r.status(now, interval)}
}

type dashboardState struct {
	host        reading[telemetry.HostInfo]
	cpuInfo     reading[telemetry.CPUInfo]
	cpu         reading[telemetry.CPUUsage]
	memory      reading[telemetry.MemoryUsage]
	disk        reading[telemetry.DiskUsage]
	network     reading[telemetry.NetworkRates]
	processes   reading[telemetry.Processes]
	temperature reading[[]telemetry.Temperature]
}

func (s *dashboardState) applyFast(msg metrics.FastSampleMsg, now time.Time) {
	if msg.CPU.HasAttempt() {
		cpuErr := msg.CPU.Err
		if cpuErr == nil && !validPercentage(msg.CPU.Usage.Usage) {
			cpuErr = errors.New("CPU usage percentage is invalid")
		}
		s.cpu.apply(msg.CPU.Usage, msg.CPU.SampledAt, msg.CPU.AttemptedAt, now, cpuErr)
	}

	if msg.Memory.HasAttempt() {
		memoryErr := msg.Memory.Err
		if memoryErr == nil && !validPercentage(msg.Memory.Usage.UsedPercent) {
			memoryErr = errors.New("memory usage percentage is invalid")
		}
		s.memory.apply(msg.Memory.Usage, msg.Memory.SampledAt, msg.Memory.AttemptedAt, now, memoryErr)
	}
}

func (s *dashboardState) applyDisk(msg metrics.DiskSampleMsg, now time.Time) {
	if msg.HasAttempt() {
		err := msg.Err
		if err == nil && !validPercentage(msg.Usage.UsedPercent) {
			err = errors.New("disk usage percentage is invalid")
		}
		s.disk.apply(msg.Usage, msg.SampledAt, msg.AttemptedAt, now, err)
	}
}

func (s *dashboardState) applyNetwork(msg metrics.NetworkSampleMsg, now time.Time) {
	if msg.HasAttempt() {
		err := msg.Err
		if err == nil && ((msg.Usage.HasReceivedRate && !validRate(msg.Usage.ReceivedBytesPerSecond)) ||
			(msg.Usage.HasSentRate && !validRate(msg.Usage.SentBytesPerSecond))) {
			err = errors.New("network rate is invalid")
		}
		s.network.apply(msg.Usage, msg.SampledAt, msg.AttemptedAt, now, err)
	}
}

func (s dashboardState) snapshot(now time.Time, intervals Intervals) dashboard.Snapshot {
	return dashboard.Snapshot{
		Host:        s.host.view(now, 0),
		CPUInfo:     s.cpuInfo.view(now, 0),
		CPU:         s.cpu.view(now, intervals.Fast),
		Memory:      s.memory.view(now, intervals.Fast),
		Disk:        s.disk.view(now, intervals.Disk),
		Network:     s.network.view(now, intervals.Network),
		Processes:   s.processes.view(now, intervals.Processes),
		Temperature: s.temperature.view(now, intervals.Temperature),
		Now:         now,
	}
}

func validRate(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func validPercentage(value float64) bool {
	return validRate(value) && value <= 100
}
