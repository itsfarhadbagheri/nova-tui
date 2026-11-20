package alerts

import (
	"log/slog"

	"github.com/itsfarhadbagheri/nova-tui/internal/events"
	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
)

type HighCPUUsage struct {
	in  <-chan any
	out chan<- any
}

func (alert *HighCPUUsage) Start() {
	for metric := range alert.in {
		switch v := metric.(type) {
		case telemetry.CPUUsage:
			if v.Usage > 90 {
				slog.Warn("high cpu usage")
				alert.out <- events.AlertReceived{Message: "high cpu usage"}
			}
		}
	}
}
