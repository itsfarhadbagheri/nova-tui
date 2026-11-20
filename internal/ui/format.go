package ui

import (
	"fmt"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

func FormatCPUUsage(value telemetry.CPUUsage) string {
	return fmt.Sprintf("%.1f%%", value.Usage)
}

func FormatMemoryUsage(value telemetry.MemoryUsage) string {
	return fmt.Sprintf("%.1f%% used", value.UsedPercent)
}

func FormatBytes(value uint64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}

	amount := float64(value)
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	for _, suffix := range units {
		amount /= unit
		if amount < unit || suffix == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", amount, suffix)
		}
	}
	return "Unavailable"
}
