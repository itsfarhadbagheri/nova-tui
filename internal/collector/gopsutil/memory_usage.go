package gopsutil

import (
	"context"
	"fmt"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/mem"
)

type MemoryUsageCollector struct {
}

func NewMemoryUsageCollector() *MemoryUsageCollector {
	return &MemoryUsageCollector{}
}

func (c *MemoryUsageCollector) Collect(ctx context.Context) (telemetry.MemoryUsage, error) {
	stat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return telemetry.MemoryUsage{}, fmt.Errorf("read virtual memory: %w", err)
	}

	return telemetry.MemoryUsage{
		Total:       stat.Total,
		Available:   stat.Available,
		Used:        stat.Used,
		UsedPercent: stat.UsedPercent,
		Free:        stat.Free,
		Active:      stat.Active,
	}, nil
}
