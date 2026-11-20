package dashboard

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui/components/systemheader"
)

const (
	wideWidth = 96
	columnGap = 2
)

type Model struct {
	styles  ui.Styles
	summary systemheader.Model
}

func New(styles ui.Styles) Model {
	return Model{styles: styles, summary: systemheader.New(styles)}
}

type Snapshot struct {
	Host        ui.Reading[telemetry.HostInfo]
	CPUInfo     ui.Reading[telemetry.CPUInfo]
	CPU         ui.Reading[telemetry.CPUUsage]
	Memory      ui.Reading[telemetry.MemoryUsage]
	Disk        ui.Reading[telemetry.DiskUsage]
	Network     ui.Reading[telemetry.NetworkRates]
	Processes   ui.Reading[telemetry.Processes]
	Temperature ui.Reading[[]telemetry.Temperature]
	Now         time.Time
}

type layout struct {
	header string
	body   []string
	rows   int
}

func (l layout) maxScroll() int {
	return max(0, len(l.body)-l.rows)
}

func (m Model) ClampScroll(width, height int, snapshot Snapshot, offset int) int {
	if offset <= 0 {
		return 0
	}
	return min(offset, m.layout(width, height, snapshot).maxScroll())
}

func (m Model) View(width, height int, snapshot Snapshot, offset int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	l := m.layout(width, height, snapshot)
	if height <= 2 {
		return l.header
	}

	offset = min(max(0, offset), l.maxScroll())
	lines := strings.Split(l.header, "\n")
	end := min(len(l.body), offset+l.rows)
	lines = append(lines, l.body[offset:end]...)
	for len(lines) < height-1 {
		lines = append(lines, "")
	}
	footer := m.styles.Label.Render("q Quit")
	if l.maxScroll() > 0 {
		scroll := "↑↓ Scroll"
		if width < 16 {
			scroll = "↑↓"
		}
		if width >= 4 {
			footer = ui.RowBetween(width, footer, m.styles.Label.Render(scroll))
		}
	}
	return ui.Constrain(strings.Join(append(lines, footer), "\n"), width, height)
}

func (m Model) layout(width, height int, snapshot Snapshot) layout {
	if width <= 0 || height <= 0 {
		return layout{}
	}
	input := systemheader.Input{
		Host: snapshot.Host, CPUInfo: snapshot.CPUInfo, CPU: snapshot.CPU,
		Memory: snapshot.Memory, Now: snapshot.Now,
	}
	if height <= 2 {
		return layout{header: m.summary.Status(width, input)}
	}

	headerRows := 1
	switch {
	case height >= 12:
		headerRows = 5
	case height >= 9:
		headerRows = 3
	case height >= 5:
		headerRows = 2
	}
	header := m.summary.View(width, headerRows, input)
	if height >= 12 {
		header += "\n" + m.styles.Border.Render(strings.Repeat("─", width))
	}
	body := strings.Split(m.panels(width, snapshot), "\n")
	return layout{header: header, body: body, rows: height - lipgloss.Height(header) - 1}
}

func (m Model) panels(width int, snapshot Snapshot) string {
	detailWidth := width
	if width >= wideWidth {
		detailWidth = (width - columnGap) * 2 / 5
	}
	details := strings.Join([]string{
		m.temperaturePanel(detailWidth, snapshot.Temperature),
		m.diskPanel(detailWidth, snapshot.Disk),
		m.networkPanel(detailWidth, snapshot.Network),
	}, "\n")
	if width < wideWidth {
		return m.processPanel(width, 0, snapshot.Processes) + "\n" + details
	}

	processWidth := width - detailWidth - columnGap
	processes := m.processPanel(processWidth, lipgloss.Height(details), snapshot.Processes)
	return lipgloss.JoinHorizontal(lipgloss.Top, processes, strings.Repeat(" ", columnGap), details)
}
