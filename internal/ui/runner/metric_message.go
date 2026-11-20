package uistarter

import (
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

type HostInfoMsg struct {
	Info        telemetry.HostInfo
	SampledAt   time.Time
	AttemptedAt time.Time
	Err         error
}

type CPUInfoMsg struct {
	Info        telemetry.CPUInfo
	SampledAt   time.Time
	AttemptedAt time.Time
	Err         error
}
