package gopsutil

import (
	"context"
	"fmt"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUInfoCollector struct {
}

func NewCPUInfoCollector() *CPUInfoCollector {
	return &CPUInfoCollector{}
}

func (c *CPUInfoCollector) Collect(ctx context.Context) (telemetry.CPUInfo, error) {
	stats, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return telemetry.CPUInfo{}, fmt.Errorf("read CPU information: %w", err)
	}

	info := telemetry.CPUInfo{}
	if len(stats) > 0 {
		info.ModelName = stats[0].Model
	}

	physicalCores, err := cpu.CountsWithContext(ctx, false)
	if err == nil && physicalCores > 0 {
		info.PhysicalCores = physicalCores
		info.HasPhysicalCores = true
	}

	return info, nil
}
