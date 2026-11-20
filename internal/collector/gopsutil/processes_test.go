package gopsutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

func TestRankProcessesUsesIntervalCPUAndProcessIdentity(t *testing.T) {
	now := time.Now()
	makeSample := func(pid int32, seconds float64, at time.Time) processSample {
		return processSample{row: telemetry.Process{PID: pid}, cpuSeconds: seconds, at: at}
	}
	previous := map[processIdentity]processSample{
		{1, 1}: makeSample(1, 10, now),
		{2, 1}: makeSample(2, 25, now),
		{3, 1}: makeSample(3, 4, now),
		{4, 1}: makeSample(4, 8, now),
		{5, 1}: makeSample(5, 13, now),
	}
	if warm := rankProcesses(nil, previous); !warm.WarmingUp || len(warm.Rows) != 0 {
		t.Fatalf("initial CPU counters should only establish a baseline: %+v", warm)
	}
	later := now.Add(4 * time.Second)
	current := map[processIdentity]processSample{
		{1, 1}: makeSample(1, 11, later),
		{2, 1}: makeSample(2, 31, later),
		{3, 1}: makeSample(3, 5, later),
		{4, 2}: makeSample(4, 20, later), // PID was reused.
		{5, 1}: makeSample(5, 13, later), // Measured zero.
		{6, 1}: makeSample(6, 15, later), // New process.
	}
	got := rankProcesses(previous, current)
	wantPIDs := []int32{2, 1, 3, 5}
	wantCPU := []float64{150, 25, 25, 0}
	if got.WarmingUp || len(got.Rows) != len(wantPIDs) {
		t.Fatalf("unexpected process rows: %+v", got)
	}
	for i, row := range got.Rows {
		if row.PID != wantPIDs[i] || row.CPUPercent != wantCPU[i] {
			t.Errorf("row %d: got PID %d CPU %g, want PID %d CPU %g", i, row.PID, row.CPUPercent, wantPIDs[i], wantCPU[i])
		}
	}
}

type fakeProcess struct {
	nameErr error
}

func (p fakeProcess) CreateTimeWithContext(context.Context) (int64, error) {
	return 10, nil
}
func (p fakeProcess) TimesWithContext(context.Context) (*cpu.TimesStat, error) {
	return &cpu.TimesStat{User: 5, System: 1}, nil
}
func (p fakeProcess) NameWithContext(context.Context) (string, error) {
	return "shell", p.nameErr
}
func (p fakeProcess) MemoryInfoWithContext(context.Context) (*process.MemoryInfoStat, error) {
	return &process.MemoryInfoStat{RSS: 1024}, nil
}

func TestReadProcessSkipsFailedReads(t *testing.T) {
	ctx := context.Background()
	if _, _, ok := readProcess(ctx, 1, fakeProcess{nameErr: errors.New("process exited")}); ok {
		t.Fatal("included an unreadable process")
	}
	id, sample, ok := readProcess(ctx, 2, fakeProcess{})
	if !ok || id.pid != 2 || id.startedAt != 10 || sample.cpuSeconds != 6 ||
		sample.row.Name != "shell" || sample.row.RSS != 1024 || sample.at.IsZero() {
		t.Fatalf("unexpected sample for a readable process: %+v %+v", id, sample)
	}
}
