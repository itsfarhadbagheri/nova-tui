package gopsutil

import (
	"context"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

type processIdentity struct {
	pid       int32
	startedAt int64
}

type processSample struct {
	row        telemetry.Process
	cpuSeconds float64
	at         time.Time
}

type ProcessesCollector struct {
	previous map[processIdentity]processSample
}

func NewProcessesCollector() *ProcessesCollector {
	return &ProcessesCollector{}
}

func (c *ProcessesCollector) Collect(ctx context.Context) (telemetry.Processes, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return telemetry.Processes{}, err
	}

	current := make(map[processIdentity]processSample, len(processes))
	for _, p := range processes {
		if err := ctx.Err(); err != nil {
			return telemetry.Processes{}, err
		}
		id, sample, ok := readProcess(ctx, p.Pid, p)
		if ok {
			current[id] = sample
		}
	}
	if err := ctx.Err(); err != nil {
		return telemetry.Processes{}, err
	}

	result := rankProcesses(c.previous, current)
	c.previous = current
	return result, nil
}

type processReader interface {
	CreateTimeWithContext(context.Context) (int64, error)
	TimesWithContext(context.Context) (*cpu.TimesStat, error)
	NameWithContext(context.Context) (string, error)
	MemoryInfoWithContext(context.Context) (*process.MemoryInfoStat, error)
}

func readProcess(ctx context.Context, pid int32, p processReader) (processIdentity, processSample, bool) {
	startedAt, err := p.CreateTimeWithContext(ctx)
	if err != nil {
		return processIdentity{}, processSample{}, false
	}
	times, err := p.TimesWithContext(ctx)
	if err != nil || times == nil {
		return processIdentity{}, processSample{}, false
	}
	at := time.Now()
	seconds := times.User + times.System
	if seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return processIdentity{}, processSample{}, false
	}
	name, err := p.NameWithContext(ctx)
	if err != nil {
		return processIdentity{}, processSample{}, false
	}
	memory, err := p.MemoryInfoWithContext(ctx)
	if err != nil || memory == nil {
		return processIdentity{}, processSample{}, false
	}
	if strings.TrimSpace(name) == "" {
		name = "Unknown"
	}
	return processIdentity{pid: pid, startedAt: startedAt}, processSample{
		row:        telemetry.Process{PID: pid, Name: name, RSS: memory.RSS},
		cpuSeconds: seconds,
		at:         at,
	}, true
}

func rankProcesses(previous, current map[processIdentity]processSample) telemetry.Processes {
	var rows []telemetry.Process
	for id, sample := range current {
		baseline, ok := previous[id]
		elapsed := sample.at.Sub(baseline.at).Seconds()
		if !ok || elapsed <= 0 || sample.cpuSeconds < baseline.cpuSeconds {
			continue
		}
		row := sample.row
		row.CPUPercent = (sample.cpuSeconds - baseline.cpuSeconds) / elapsed * 100
		rows = append(rows, row)
	}
	slices.SortFunc(rows, func(a, b telemetry.Process) int {
		if a.CPUPercent > b.CPUPercent {
			return -1
		}
		if a.CPUPercent < b.CPUPercent {
			return 1
		}
		// PIDs are positive int32 values.
		return int(a.PID) - int(b.PID)
	})
	return telemetry.Processes{Rows: rows, WarmingUp: len(rows) == 0 && len(current) > 0}
}
