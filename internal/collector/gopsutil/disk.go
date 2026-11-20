package gopsutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/disk"
)

type DiskUsageCollector struct{}

func NewDiskUsageCollector() *DiskUsageCollector {
	return &DiskUsageCollector{}
}

func (c *DiskUsageCollector) Collect(ctx context.Context) (telemetry.DiskUsage, error) {
	if err := ctx.Err(); err != nil {
		return telemetry.DiskUsage{}, err
	}
	path, err := rootPath()
	if err != nil {
		return telemetry.DiskUsage{}, err
	}

	usage, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return telemetry.DiskUsage{}, fmt.Errorf("read disk usage for %q: %w", path, err)
	}
	return telemetry.DiskUsage{
		Path:        usage.Path,
		Total:       usage.Total,
		Free:        usage.Free,
		Used:        usage.Used,
		UsedPercent: usage.UsedPercent,
	}, nil
}

func rootPath() (string, error) {
	if runtime.GOOS != "windows" {
		return string(filepath.Separator), nil
	}
	drive := os.Getenv("SystemDrive")
	if drive == "" {
		return "", fmt.Errorf("resolve Windows system drive: SystemDrive is empty")
	}
	return filepath.Clean(drive + string(filepath.Separator)), nil
}
