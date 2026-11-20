package systemheader

import (
	"testing"
	"time"
)

func TestFormatUptimeClampsFutureBootTime(t *testing.T) {
	now := time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)
	bootTime := now.Add(time.Hour)

	if got := formatUptime(now, bootTime); got != "0s" {
		t.Fatalf("future boot time should produce zero uptime, got %q", got)
	}
}

func TestFormatUptimeShowsDashWhenBootTimeIsUnavailable(t *testing.T) {
	if got := formatUptime(time.Now(), time.Time{}); got != "—" {
		t.Fatalf("missing boot time should produce a dash, got %q", got)
	}
}
