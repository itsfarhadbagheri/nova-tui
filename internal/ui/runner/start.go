package uistarter

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itsfarhadbagheri/nova-tui/internal/metrics"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui/pages/dashboard"
)

type model struct {
	dashboard dashboard.Model
	state     dashboardState
	intervals Intervals

	width  int
	height int
	scroll int
}

// Intervals are the polling intervals used to determine reading freshness.
type Intervals struct {
	Fast, Disk, Network    time.Duration
	Processes, Temperature time.Duration
}

func NewModel(intervals Intervals) model {
	theme := ui.DefaultTheme()
	styles := ui.NewStyles(theme)

	return model{
		dashboard: dashboard.New(styles),
		intervals: intervals,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	now := time.Now()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			m.scroll--
		case "down":
			m.scroll++
		}

	case metrics.FastSampleMsg:
		m.state.applyFast(msg, now)
	case metrics.DiskSampleMsg:
		m.state.applyDisk(msg, now)
	case metrics.NetworkSampleMsg:
		m.state.applyNetwork(msg, now)
	case metrics.ProcessesSampleMsg:
		if msg.HasAttempt() {
			m.state.processes.apply(msg.Usage, msg.SampledAt, msg.AttemptedAt, now, msg.Err)
		}
	case metrics.TemperatureSampleMsg:
		if msg.HasAttempt() {
			m.state.temperature.apply(msg.Usage, msg.SampledAt, msg.AttemptedAt, now, msg.Err)
		}
	case HostInfoMsg:
		m.state.host.apply(msg.Info, msg.SampledAt, msg.AttemptedAt, now, msg.Err)

	case CPUInfoMsg:
		m.state.cpuInfo.apply(msg.Info, msg.SampledAt, msg.AttemptedAt, now, msg.Err)
	}

	if m.scroll != 0 {
		m.scroll = m.dashboard.ClampScroll(m.width, m.height, m.state.snapshot(now, m.intervals), m.scroll)
	}
	return m, nil
}

func (m model) View() string {
	return m.viewAt(time.Now())
}

func (m model) viewAt(now time.Time) string {
	return m.dashboard.View(m.width, m.height, m.state.snapshot(now, m.intervals), m.scroll)
}
