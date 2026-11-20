package gopsutil

import (
	"context"
	"fmt"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/host"
)

type HostInfoCollector struct {
}

func NewHostInfoCollector() *HostInfoCollector {
	return &HostInfoCollector{}
}

func (c *HostInfoCollector) Collect(ctx context.Context) (telemetry.HostInfo, error) {
	stat, err := host.InfoWithContext(ctx)
	if err != nil {
		return telemetry.HostInfo{}, fmt.Errorf("read host information: %w", err)
	}

	var bootTime time.Time
	if stat.BootTime > 0 {
		bootTime = time.Unix(int64(stat.BootTime), 0)
	}

	return telemetry.HostInfo{
		Hostname:        stat.Hostname,
		BootTime:        bootTime,
		Procs:           stat.Procs,
		OS:              stat.OS,
		Platform:        stat.Platform,
		PlatformFamily:  stat.PlatformFamily,
		PlatformVersion: stat.PlatformVersion,
		KernelVersion:   stat.KernelVersion,
		KernelArch:      stat.KernelArch,
	}, nil
}
