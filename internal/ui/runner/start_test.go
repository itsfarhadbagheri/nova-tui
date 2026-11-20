package uistarter

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itsfarhadbagheri/nova-tui/internal/metrics"
	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/itsfarhadbagheri/nova-tui/internal/ui"
)

func TestCPUUsageAppearsOnceInOverview(t *testing.T) {
	uiModel := sizedModel(t)
	now := time.Now()
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			Usage:       telemetry.CPUUsage{Usage: 42.5},
			SampledAt:   now,
			AttemptedAt: now,
		},
	})

	view := uiModel.View()
	if count := strings.Count(view, "42.5%"); count != 1 {
		t.Fatalf("CPU sample should appear once in the overview; found %d occurrences in:\n%s", count, view)
	}
}

func TestReadingWaitingZeroFailureRecoveryAndStale(t *testing.T) {
	uiModel := sizedModel(t)
	if view := uiModel.View(); !strings.Contains(view, "Waiting for sample") {
		t.Fatalf("initial readings should wait for their first sample; view:\n%s", view)
	}

	now := time.Now()
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			Usage:       telemetry.CPUUsage{Usage: 0},
			SampledAt:   now,
			AttemptedAt: now,
		},
	})
	if view := uiModel.View(); !strings.Contains(view, "0.0%") {
		t.Fatalf("a measured zero should be rendered as valid; view:\n%s", view)
	}

	failureAt := now.Add(time.Second)
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			AttemptedAt: failureAt,
			Err:         errors.New("collector failed"),
		},
	})
	if view := uiModel.View(); !strings.Contains(view, "0.0% (stale)") {
		t.Fatalf("a failure after success should preserve and mark the last value stale; view:\n%s", view)
	}

	recoveryAt := failureAt.Add(time.Second)
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			Usage:       telemetry.CPUUsage{Usage: 17.3},
			SampledAt:   recoveryAt,
			AttemptedAt: recoveryAt,
		},
	})
	if view := uiModel.View(); !strings.Contains(view, "17.3%") || strings.Contains(view, "17.3% (stale)") {
		t.Fatalf("a successful sample should recover from stale state; view:\n%s", view)
	}
}

func TestFirstFailureIsUnavailableAndMetricsStayIndependent(t *testing.T) {
	uiModel := sizedModel(t)
	now := time.Now()
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			AttemptedAt: now,
			Err:         errors.New("CPU failed"),
		},
		Memory: metrics.MemoryUsageSample{
			Usage: telemetry.MemoryUsage{
				Available:   2 * 1024 * 1024,
				UsedPercent: 35,
			},
			SampledAt:   now,
			AttemptedAt: now,
		},
	})

	view := uiModel.View()
	if !strings.Contains(view, "Unavailable") {
		t.Fatalf("a first failed sample should be unavailable; view:\n%s", view)
	}
	if !strings.Contains(view, "35.0% used") || !strings.Contains(view, "2.0 MiB available") {
		t.Fatalf("a memory success should remain visible after an independent CPU failure; view:\n%s", view)
	}
}

func TestDiskAndInterfaceTrafficRenderTheirScopeAndRates(t *testing.T) {
	uiModel := sizedModel(t)
	now := time.Now()
	uiModel = updateModel(uiModel, metrics.DiskSampleMsg{
		Usage: telemetry.DiskUsage{
			Path:        "/data",
			Free:        8 * 1024 * 1024 * 1024,
			UsedPercent: 71.2,
		},
		SampledAt:   now,
		AttemptedAt: now,
	})
	uiModel = updateModel(uiModel, metrics.NetworkSampleMsg{
		Usage: telemetry.NetworkRates{
			ReceivedBytesPerSecond: 1024,
			SentBytesPerSecond:     0,
			HasReceivedRate:        true,
			HasSentRate:            true,
		},
		SampledAt:   now,
		AttemptedAt: now,
	})

	view := uiModel.View()
	for _, expected := range []string{"/data", "Interface traffic", "1.0 KiB/s", "0.0 B/s"} {
		if !strings.Contains(view, expected) {
			t.Errorf("dashboard should include %q, view:\n%s", expected, view)
		}
	}

	failureAt := now.Add(time.Second)
	uiModel = updateModel(uiModel, metrics.DiskSampleMsg{AttemptedAt: failureAt, Err: errors.New("disk failed")})
	uiModel = updateModel(uiModel, metrics.NetworkSampleMsg{AttemptedAt: failureAt, Err: errors.New("network failed")})
	if view := uiModel.View(); !strings.Contains(view, "(stale)") {
		t.Fatalf("failed slow polls should keep prior readings and mark them stale; view:\n%s", view)
	}
}

func TestPercentageValidation(t *testing.T) {
	for _, value := range []float64{0, 100, 23.4} {
		if !validPercentage(value) {
			t.Errorf("valid percentage %v was rejected", value)
		}
	}
	for _, value := range []float64{-0.1, 100.1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if validPercentage(value) {
			t.Errorf("invalid percentage %v was accepted", value)
		}
	}
}

func TestReadingBecomesStaleAfterTwiceInterval(t *testing.T) {
	sampledAt := time.Date(2026, time.November, 18, 0, 0, 0, 0, time.UTC)
	var cpu reading[telemetry.CPUUsage]
	cpu.apply(telemetry.CPUUsage{Usage: 1}, sampledAt, sampledAt, sampledAt, nil)

	if status := cpu.status(sampledAt.Add(2*time.Second), time.Second); status != ui.ReadingStatusCurrent {
		t.Fatalf("reading at exactly twice the interval should remain current, got %v", status)
	}
	if status := cpu.status(sampledAt.Add(2*time.Second+time.Nanosecond), time.Second); status != ui.ReadingStatusStale {
		t.Fatalf("reading older than twice the interval should be stale, got %v", status)
	}
}

func TestViewFitsResponsiveTerminalBounds(t *testing.T) {
	uiModel := newTestModel()
	now := time.Now()
	uiModel = updateModel(uiModel, HostInfoMsg{
		Info: telemetry.HostInfo{
			Hostname:      "界界界界界界界界界界-long-hostname",
			OS:            "long operating system name",
			KernelVersion: "kernel-version-that-is-longer-than-the-terminal",
			BootTime:      now.Add(-time.Hour),
		},
		SampledAt:   now,
		AttemptedAt: now,
	})
	uiModel = updateModel(uiModel, CPUInfoMsg{
		Info:        telemetry.CPUInfo{ModelName: "wide界 CPU model name that is intentionally long", PhysicalCores: 8, HasPhysicalCores: true},
		SampledAt:   now,
		AttemptedAt: now,
	})
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			Usage:       telemetry.CPUUsage{Usage: 42.5},
			SampledAt:   now,
			AttemptedAt: now,
		},
		Memory: metrics.MemoryUsageSample{
			Usage:       telemetry.MemoryUsage{Available: 6 * 1024 * 1024 * 1024, UsedPercent: 63.2},
			SampledAt:   now,
			AttemptedAt: now,
		},
	})
	uiModel = updateModel(uiModel, metrics.DiskSampleMsg{
		Usage:       telemetry.DiskUsage{Path: "/界界界界界界界界界", Free: 4 * 1024 * 1024, UsedPercent: 25},
		SampledAt:   now,
		AttemptedAt: now,
	})
	uiModel = updateModel(uiModel, metrics.NetworkSampleMsg{
		Usage:       telemetry.NetworkRates{ReceivedBytesPerSecond: 1.5, SentBytesPerSecond: 2.5, HasReceivedRate: true, HasSentRate: true},
		SampledAt:   now,
		AttemptedAt: now,
	})

	uiModel = updateModel(uiModel, metrics.ProcessesSampleMsg{
		Usage: telemetry.Processes{Rows: []telemetry.Process{
			{PID: 123456, Name: "界界界界界界界界\nlong process name", CPUPercent: 125, RSS: 1024},
		}},
		SampledAt: now,
	})
	uiModel = updateModel(uiModel, metrics.TemperatureSampleMsg{
		Usage: []telemetry.Temperature{
			{Name: "界界界界界界界界\tvery long sensor name", Celsius: 55},
		},
		SampledAt: now,
	})

	for _, width := range []int{1, 15, 16, 49, 80, 95, 96, 120} {
		for _, height := range []int{1, 2, 3, 8, 16, 32} {
			sized := updateModel(uiModel, tea.WindowSizeMsg{Width: width, Height: height})
			view := sized.View()
			if got := lipgloss.Height(view); got > height {
				t.Errorf("view is %d rows at %dx%d", got, width, height)
			}
			for row, line := range strings.Split(view, "\n") {
				if got := lipgloss.Width(line); got > width {
					t.Errorf("row %d is %d cells at %dx%d", row, got, width, height)
				}
			}
			if height < 3 && lipgloss.Height(view) != 1 {
				t.Errorf("height %d should render one status row, got %d", height, lipgloss.Height(view))
			}
		}
	}
}

func TestCompactViewRetainsCPUAndMemoryRows(t *testing.T) {
	uiModel := sizedModel(t)
	now := time.Now()
	uiModel = updateModel(uiModel, metrics.FastSampleMsg{
		CPU: metrics.CPUUsageSample{
			Usage:       telemetry.CPUUsage{Usage: 42.5},
			SampledAt:   now,
			AttemptedAt: now,
		},
		Memory: metrics.MemoryUsageSample{
			Usage:       telemetry.MemoryUsage{Available: 3 * 1024 * 1024, UsedPercent: 63.2},
			SampledAt:   now,
			AttemptedAt: now,
		},
	})
	uiModel = updateModel(uiModel, tea.WindowSizeMsg{Width: 49, Height: 4})
	view := uiModel.View()
	if !strings.Contains(view, "CPU") || !strings.Contains(view, "Memory") {
		t.Fatalf("compact view should show CPU and memory rows, got:\n%s", view)
	}
}

func sizedModel(t *testing.T) model {
	t.Helper()
	uiModel := newTestModel()
	return updateModel(uiModel, tea.WindowSizeMsg{Width: 100, Height: 40})
}

func updateModel(current model, message tea.Msg) model {
	updated, _ := current.Update(message)
	return updated.(model)
}

func newTestModel() model {
	return NewModel(Intervals{
		Fast:        metrics.FastInterval,
		Disk:        metrics.DiskInterval,
		Network:     metrics.NetworkInterval,
		Processes:   metrics.ProcessesInterval,
		Temperature: metrics.TemperatureInterval,
	})
}

func TestResponsiveLayoutTransitions(t *testing.T) {
	for _, tc := range []struct {
		width, height int
		wide, scroll  bool
	}{
		{95, 25, false, false},
		{96, 25, true, false},
		{95, 6, false, true},
		{96, 6, true, true},
	} {
		m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: tc.width, Height: tc.height})
		view := m.viewAt(time.Time{})
		var wide bool
		for _, line := range strings.Split(view, "\n") {
			if strings.Contains(line, "Top processes") && strings.Contains(line, "Temperature") {
				wide = true
			}
		}
		if wide != tc.wide || strings.Contains(view, "↑↓ Scroll") != tc.scroll {
			t.Errorf("%dx%d: want wide=%v, scroll=%v, view:\n%s", tc.width, tc.height, tc.wide, tc.scroll, view)
		}
	}
}

func TestScrollingKeepsOverviewVisibleAndClampsOnResize(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 60, Height: 6})
	now := time.Now()
	original := m.viewAt(now)
	pinned := strings.Split(original, "\n")[:2]
	found := make(map[string]bool)
	for i := 0; i < 30; i++ {
		view := m.viewAt(now)
		for _, title := range []string{"Top processes", "Temperature", "Disk", "Interface traffic", "TX:"} {
			found[title] = found[title] || strings.Contains(view, title)
		}
		if !strings.HasPrefix(view, strings.Join(pinned, "\n")+"\n") {
			t.Fatalf("scroll displaced the overview:\n%s", view)
		}
		m = updateModel(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	for title, visible := range found {
		if !visible {
			t.Errorf("%q was unreachable", title)
		}
	}
	if bottom := updateModel(m, tea.KeyMsg{Type: tea.KeyDown}); bottom.scroll != m.scroll || m.scroll <= 0 {
		t.Fatalf("down should clamp at the bottom, got %d then %d", m.scroll, bottom.scroll)
	}
	if up := updateModel(m, tea.KeyMsg{Type: tea.KeyUp}); up.scroll != m.scroll-1 {
		t.Fatalf("up should reveal the previous line, got offset %d", up.scroll)
	}
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.scroll != 0 || strings.Contains(m.viewAt(now), "↑↓ Scroll") {
		t.Fatalf("resize should reset offset when the body fits, got %d:\n%s", m.scroll, m.viewAt(now))
	}
}

func TestProcessAndTemperaturePanels(t *testing.T) {
	m := sizedModel(t)
	now := time.Now()
	m = updateModel(m, metrics.ProcessesSampleMsg{Usage: telemetry.Processes{WarmingUp: true}, SampledAt: now})
	m = updateModel(m, metrics.TemperatureSampleMsg{SampledAt: now})
	for _, text := range []string{"Waiting for CPU samples", "No temperature sensors found"} {
		if !strings.Contains(m.viewAt(now), text) {
			t.Errorf("missing %q:\n%s", text, m.viewAt(now))
		}
	}

	var processes []telemetry.Process
	for pid := range 9 {
		processes = append(processes, telemetry.Process{PID: int32(pid + 1), Name: "測試-process", CPUPercent: 125, RSS: 1024})
	}
	processes[8].Name = "ninth process"
	m = updateModel(m, metrics.ProcessesSampleMsg{
		Usage: telemetry.Processes{Rows: processes}, SampledAt: now,
	})
	m = updateModel(m, metrics.TemperatureSampleMsg{
		Usage: []telemetry.Temperature{
			{Name: "CPU package", Celsius: 61},
			{Name: "Drive", Celsius: 52.5},
			{Name: "Board", Celsius: 40},
			{Name: "fourth sensor", Celsius: 38},
		},
		SampledAt: now,
	})
	view := m.viewAt(now)
	for _, text := range []string{"PID", "RSS", "125.0%", "1.0 KiB", "CPU package", "61.0 °C", "52.5 °C", "40.0 °C"} {
		if !strings.Contains(view, text) {
			t.Errorf("missing %q:\n%s", text, view)
		}
	}
	if strings.Count(view, "測試-process") != 8 || strings.Contains(view, "ninth process") || strings.Contains(view, "fourth sensor") {
		t.Fatalf("expected eight processes and three sensors:\n%s", view)
	}
	m = updateModel(m, metrics.TemperatureSampleMsg{Err: errors.New("sensor poll failed")})
	if view := m.viewAt(now); !strings.Contains(view, "Readings are stale") || !strings.Contains(view, "61.0 °C") {
		t.Fatalf("temperature failure should retain the stale value:\n%s", view)
	}
	m = updateModel(m, metrics.TemperatureSampleMsg{SampledAt: now.Add(time.Second)})
	if view := m.viewAt(now.Add(time.Second)); strings.Contains(view, "Readings are stale") || !strings.Contains(view, "No temperature sensors found") {
		t.Fatalf("successful poll should recover from stale state:\n%s", view)
	}
}

func TestQuitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyCtrlC},
	} {
		_, cmd := newTestModel().Update(key)
		if cmd == nil {
			t.Errorf("%s did not return a command", key)
			continue
		}
		if msg := cmd(); msg != tea.Quit() {
			t.Errorf("%s returned %T, want quit", key, msg)
		}
	}
}

func TestNetworkDirectionsRenderIndependently(t *testing.T) {
	now := time.Date(2026, time.November, 20, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name           string
		rates          telemetry.NetworkRates
		wantRX, wantTX string
	}{
		{"baseline", telemetry.NetworkRates{}, "Unavailable", "Unavailable"},
		{"receive only", telemetry.NetworkRates{
			HasReceivedRate: true, ReceivedBytesPerSecond: 1024,
			SentBytesPerSecond: math.NaN(),
		}, "1.0 KiB/s", "Unavailable"},
		{"send only", telemetry.NetworkRates{
			HasSentRate: true, SentBytesPerSecond: 0,
			ReceivedBytesPerSecond: math.NaN(),
		}, "Unavailable", "0.0 B/s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := updateModel(sizedModel(t), metrics.NetworkSampleMsg{Usage: tc.rates, SampledAt: now})
			for _, expectation := range []struct {
				at     time.Time
				suffix string
			}{
				{now.Add(2 * metrics.NetworkInterval), ""},
				{now.Add(2*metrics.NetworkInterval + time.Nanosecond), " (stale)"},
			} {
				view := m.viewAt(expectation.at)
				for _, text := range []string{"RX: " + tc.wantRX, "TX: " + tc.wantTX} {
					if !strings.Contains(view, text+expectation.suffix) ||
						(expectation.suffix == "" && strings.Contains(view, text+" (stale)")) {
						t.Errorf("expected %q at age %s, view:\n%s", text+expectation.suffix, expectation.at.Sub(now), view)
					}
				}
			}
		})
	}
}
