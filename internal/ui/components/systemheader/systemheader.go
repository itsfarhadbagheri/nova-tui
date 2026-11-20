package systemheader

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui"
)

type Model struct {
	styles ui.Styles
}

func New(styles ui.Styles) Model {
	return Model{styles: styles}
}

type Input struct {
	Host    ui.Reading[telemetry.HostInfo]
	CPUInfo ui.Reading[telemetry.CPUInfo]
	CPU     ui.Reading[telemetry.CPUUsage]
	Memory  ui.Reading[telemetry.MemoryUsage]
	Now     time.Time
}

func (m Model) Status(width int, input Input) string {
	cpu := ui.ReadingText(input.CPU, ui.FormatCPUUsage)
	memory := ui.ReadingText(input.Memory, ui.FormatMemoryUsage)
	line := m.styles.Accent.Render("NOVA") + "  " + m.styles.Label.Render("CPU ") +
		m.styles.Value.Render(cpu) + " · " + m.styles.Label.Render("Memory ") + m.styles.Value.Render(memory)
	return ui.TruncateCells(line, width)
}

func (m Model) View(width, height int, input Input) string {
	if height <= 1 {
		return m.Status(width, input)
	}
	meters := m.meters(width, input)
	if height == 2 {
		return strings.Join(meters, "\n")
	}

	hostname := ui.ReadingText(input.Host, func(host telemetry.HostInfo) string {
		if name := ui.SingleLine(host.Hostname); name != "" {
			return name
		}
		return "Unavailable"
	})
	identity := m.styles.Accent.Render("NOVA") + "  " + m.styles.Value.Render(hostname)
	if input.Host.HasValue() && !input.Host.Value.BootTime.IsZero() {
		uptime := ui.ReadingText(input.Host, func(host telemetry.HostInfo) string {
			return "Up " + formatUptime(input.Now, host.BootTime)
		})
		if width >= 24 {
			identity = ui.RowBetween(width, identity, m.styles.Label.Render(uptime))
		}
	}
	lines := []string{identity}
	if height >= 4 {
		os := ui.ReadingText(input.Host, func(host telemetry.HostInfo) string {
			text := strings.TrimSpace(ui.SingleLine(host.OS) + " " + ui.SingleLine(host.KernelVersion))
			if text == "" {
				return "Unavailable"
			}
			return text
		})
		hardware := ui.ReadingText(input.CPUInfo, func(info telemetry.CPUInfo) string {
			detailWidth := width
			if input.CPUInfo.Status == ui.ReadingStatusStale {
				detailWidth -= len(" (stale)")
			}
			return formatCPUInfo(info, detailWidth)
		})
		details := os + " · " + hardware
		if lipgloss.Width(details) > width && height >= 5 {
			lines = append(lines, m.styles.Label.Render(os), m.styles.Label.Render(hardware))
		} else {
			lines = append(lines, m.styles.Label.Render(details))
		}
	}
	return ui.Constrain(strings.Join(append(lines, meters...), "\n"), width, height)
}

func (m Model) meters(width int, input Input) []string {
	cpu := ui.ReadingText(input.CPU, ui.FormatCPUUsage)
	memory := ui.ReadingText(input.Memory, ui.FormatMemoryUsage)
	extra := ""
	if input.Memory.HasValue() {
		extra = ui.FormatBytes(input.Memory.Value.Available) + " available"
	}
	valueWidth := max(lipgloss.Width(cpu), lipgloss.Width(memory))
	if width >= 40 {
		valueWidth = max(18, valueWidth)
	}
	barWidth := min(24, width-8-valueWidth-2)
	showExtra := extra != "" && width >= 8+valueWidth+8+lipgloss.Width(extra)+4
	if showExtra {
		barWidth = min(barWidth, width-8-valueWidth-lipgloss.Width(extra)-4)
	}
	cpuLine := m.meter("CPU", cpu, "", valueWidth, barWidth, input.CPU.HasValue(), input.CPU.Value.Usage)
	memoryExtra := ""
	if showExtra {
		memoryExtra = extra
	} else if extra != "" && lipgloss.Width("Memory  "+memory+" · "+extra) <= width {
		return []string{
			m.meter("CPU", cpu, "", valueWidth, 0, false, 0),
			m.meter("Memory", memory+" · "+extra, "", valueWidth, 0, false, 0),
		}
	}
	return []string{
		ui.TruncateCells(cpuLine, width),
		ui.TruncateCells(m.meter("Memory", memory, memoryExtra, valueWidth, barWidth, input.Memory.HasValue(), input.Memory.Value.UsedPercent), width),
	}
}

func (m Model) meter(label, value, detail string, valueWidth, barWidth int, hasValue bool, percent float64) string {
	line := m.styles.Label.Render(label+strings.Repeat(" ", 8-len(label))) + m.styles.Value.Render(value)
	if hasValue && barWidth >= 4 {
		line += strings.Repeat(" ", valueWidth-lipgloss.Width(value)+2) + ui.PercentageBar(m.styles, percent, barWidth)
	}
	if detail != "" {
		line += "  " + m.styles.Label.Render(detail)
	}
	return line
}

func formatCPUInfo(info telemetry.CPUInfo, width int) string {
	cores := "Physical cores unavailable"
	if info.HasPhysicalCores {
		cores = fmt.Sprintf("%d physical cores", info.PhysicalCores)
	}
	separator := " · "
	model := ui.TruncateCells(ui.SingleLine(info.ModelName), width-lipgloss.Width(separator+cores))
	if model == "" {
		return cores
	}
	return model + separator + cores
}

func formatUptime(now, bootTime time.Time) string {
	if bootTime.IsZero() {
		return "—"
	}
	duration := max(0, now.Sub(bootTime))
	days := duration / (24 * time.Hour)
	duration %= 24 * time.Hour
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, duration/time.Hour)
	}
	if duration >= time.Hour {
		return fmt.Sprintf("%dh %dm", duration/time.Hour, duration%time.Hour/time.Minute)
	}
	return duration.Truncate(time.Second).String()
}
