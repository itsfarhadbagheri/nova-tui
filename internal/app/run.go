package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itsfarhadbagheri/nova-tui/internal/collector/gopsutil"
	"github.com/itsfarhadbagheri/nova-tui/internal/metrics"
	uistarter "github.com/itsfarhadbagheri/nova-tui/internal/ui/runner"
)

func Run(ctx context.Context) error {
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	intervals := uistarter.Intervals{
		Fast:        metrics.FastInterval,
		Disk:        metrics.DiskInterval,
		Network:     metrics.NetworkInterval,
		Processes:   metrics.ProcessesInterval,
		Temperature: metrics.TemperatureInterval,
	}
	p := tea.NewProgram(
		uistarter.NewModel(intervals),
		tea.WithAltScreen(),
		tea.WithoutSignalHandler(), // main cancels the context on SIGINT and SIGTERM.
		tea.WithContext(workerCtx),
	)

	var workers sync.WaitGroup
	workers.Add(7)
	go func() {
		defer workers.Done()
		metrics.RunFast(
			workerCtx,
			intervals.Fast,
			gopsutil.NewCPUUsageCollector(),
			gopsutil.NewMemoryUsageCollector(),
			func(sample metrics.FastSampleMsg) { p.Send(sample) },
		)
	}()
	go func() {
		defer workers.Done()
		metrics.RunDisk(
			workerCtx,
			intervals.Disk,
			gopsutil.NewDiskUsageCollector(),
			func(sample metrics.DiskSampleMsg) { p.Send(sample) },
		)
	}()
	go func() {
		defer workers.Done()
		collector := gopsutil.NewNetworkUsageCollector()
		metrics.RunNetwork(
			workerCtx,
			intervals.Network,
			collector,
			func(sample metrics.NetworkSampleMsg) { p.Send(sample) },
		)
	}()

	go func() {
		defer workers.Done()
		metrics.RunSample(workerCtx, intervals.Processes, gopsutil.NewProcessesCollector(),
			func(sample metrics.ProcessesSampleMsg) { p.Send(sample) })
	}()
	go func() {
		defer workers.Done()
		metrics.RunSample(workerCtx, intervals.Temperature, gopsutil.NewTemperatureCollector(),
			func(sample metrics.TemperatureSampleMsg) { p.Send(sample) })
	}()

	go func() {
		defer workers.Done()
		attemptedAt := time.Now()
		info, err := gopsutil.NewHostInfoCollector().Collect(workerCtx)
		if workerCtx.Err() != nil {
			return
		}
		sample := uistarter.HostInfoMsg{Info: info, AttemptedAt: attemptedAt, Err: err}
		if err == nil {
			sample.SampledAt = time.Now()
		}
		p.Send(sample)
	}()

	go func() {
		defer workers.Done()
		attemptedAt := time.Now()
		info, err := gopsutil.NewCPUInfoCollector().Collect(workerCtx)
		if workerCtx.Err() != nil {
			return
		}
		sample := uistarter.CPUInfoMsg{Info: info, AttemptedAt: attemptedAt, Err: err}
		if err == nil {
			sample.SampledAt = time.Now()
		}
		p.Send(sample)
	}()

	_, runErr := p.Run()
	cancel()
	workers.Wait()
	if runErr != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("bubble tea runtime error: %w", runErr)
	}

	return nil
}
