package gopsutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUUsageCollector struct {
}

func NewCPUUsageCollector() *CPUUsageCollector {
	return &CPUUsageCollector{}
}

func (c *CPUUsageCollector) Collect(ctx context.Context) (telemetry.CPUUsage, error) {
	usage, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return telemetry.CPUUsage{}, fmt.Errorf("read aggregate CPU usage: %w", err)
	}
	if len(usage) == 0 {
		return telemetry.CPUUsage{}, errors.New("aggregate CPU sample is empty")
	}

	return telemetry.CPUUsage{Usage: usage[0]}, nil
}
