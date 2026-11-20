package dashboard

import (
	"fmt"
	"strconv"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui"
)

func (m Model) processPanel(width, height int, reading ui.Reading[telemetry.Processes]) string {
	var body []string
	switch {
	case !reading.HasValue():
		body = []string{ui.ReadingText(reading, nil)}
	case reading.Value.WarmingUp:
		body = []string{ui.ReadingText(reading, func(telemetry.Processes) string { return "Waiting for CPU samples" })}
	case len(reading.Value.Rows) == 0:
		body = []string{ui.ReadingText(reading, func(telemetry.Processes) string { return "No readable processes" })}
	default:
		body = m.processTable(ui.PanelBodyWidth(width), reading.Value.Rows)
		if reading.Status == ui.ReadingStatusStale {
			body = append([]string{m.styles.Label.Render("Readings are stale")}, body...)
		}
	}
	return ui.Panel(m.styles, width, height, "Top processes", body)
}

func (m Model) processTable(width int, processes []telemetry.Process) []string {
	processes = processes[:min(8, len(processes))]
	pidWidth, cpuWidth, rssWidth := 3, 3, 3
	for _, p := range processes {
		pidWidth = max(pidWidth, len(strconv.FormatInt(int64(p.PID), 10)))
		cpuWidth = max(cpuWidth, len(processCPU(p)))
		rssWidth = max(rssWidth, len(ui.FormatBytes(p.RSS)))
	}
	showPID := width >= pidWidth+cpuWidth+7
	showRSS := width >= pidWidth+cpuWidth+rssWidth+10
	showCPU := width >= cpuWidth+4
	nameWidth := width
	var columns []ui.Column
	if showPID {
		columns = append(columns, ui.Column{Title: "PID", Width: pidWidth, Right: true})
		nameWidth -= pidWidth + 1
	}
	if showCPU {
		nameWidth -= cpuWidth + 1
	}
	if showRSS {
		nameWidth -= rssWidth + 1
	}
	columns = append(columns, ui.Column{Title: "Process", Width: nameWidth})
	if showCPU {
		columns = append(columns, ui.Column{Title: "CPU", Width: cpuWidth, Right: true})
	}
	if showRSS {
		columns = append(columns, ui.Column{Title: "RSS", Width: rssWidth, Right: true})
	}

	rows := make([][]string, 0, len(processes))
	for _, p := range processes {
		var row []string
		if showPID {
			row = append(row, strconv.FormatInt(int64(p.PID), 10))
		}
		row = append(row, ui.SingleLine(p.Name))
		if showCPU {
			row = append(row, processCPU(p))
		}
		if showRSS {
			row = append(row, ui.FormatBytes(p.RSS))
		}
		rows = append(rows, row)
	}
	return ui.Table(m.styles, columns, rows)
}

func processCPU(p telemetry.Process) string {
	return fmt.Sprintf("%.1f%%", p.CPUPercent)
}

func (m Model) temperaturePanel(width int, reading ui.Reading[[]telemetry.Temperature]) string {
	var body []string
	if !reading.HasValue() {
		body = []string{ui.ReadingText(reading, nil)}
	} else {
		if reading.Status == ui.ReadingStatusStale {
			body = append(body, m.styles.Label.Render("Readings are stale"))
		}
		if len(reading.Value) == 0 {
			body = append(body, "No temperature sensors found")
		}
		for _, sensor := range reading.Value[:min(3, len(reading.Value))] {
			value := m.styles.Value.Render(fmt.Sprintf("%.1f °C", sensor.Celsius))
			body = append(body, ui.RowBetween(ui.PanelBodyWidth(width), ui.SingleLine(sensor.Name), value))
		}
	}
	return ui.Panel(m.styles, width, 0, "Temperature", body)
}

func (m Model) diskPanel(width int, reading ui.Reading[telemetry.DiskUsage]) string {
	title := "Disk"
	if path := ui.SingleLine(reading.Value.Path); reading.HasValue() && path != "" {
		title += " · " + path
	}
	usage := ui.ReadingText(reading, func(disk telemetry.DiskUsage) string {
		return fmt.Sprintf("%.1f%% used", disk.UsedPercent)
	})
	body := []string{m.styles.Value.Render(usage)}
	if reading.HasValue() {
		free := ui.FormatBytes(reading.Value.Free) + " free"
		barWidth := min(20, ui.PanelBodyWidth(width))
		body = append(body, m.styles.Label.Render(free))
		if barWidth >= 8 {
			body = append(body, ui.PercentageBar(m.styles, reading.Value.UsedPercent, barWidth))
		}
	}
	return ui.Panel(m.styles, width, 0, title, body)
}

func (m Model) networkPanel(width int, reading ui.Reading[telemetry.NetworkRates]) string {
	rx := networkDirection(reading, true)
	tx := networkDirection(reading, false)
	body := []string{
		m.styles.Label.Render("RX: ") + m.styles.Value.Render(rx),
		m.styles.Label.Render("TX: ") + m.styles.Value.Render(tx),
	}
	return ui.Panel(m.styles, width, 0, "Interface traffic", body)
}

func networkDirection(reading ui.Reading[telemetry.NetworkRates], receive bool) string {
	return ui.ReadingText(reading, func(rates telemetry.NetworkRates) string {
		if receive && rates.HasReceivedRate {
			return formatRate(rates.ReceivedBytesPerSecond)
		}
		if !receive && rates.HasSentRate {
			return formatRate(rates.SentBytesPerSecond)
		}
		return "Unavailable"
	})
}

func formatRate(value float64) string {
	if value < 1024 {
		return fmt.Sprintf("%.1f B/s", value)
	}
	units := []string{"KiB/s", "MiB/s", "GiB/s", "TiB/s", "PiB/s"}
	for _, suffix := range units {
		value /= 1024
		if value < 1024 || suffix == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return "Unavailable"
}
