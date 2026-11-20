package telemetry

type Processes struct {
	Rows      []Process
	WarmingUp bool
}

type Process struct {
	PID        int32
	Name       string
	CPUPercent float64
	RSS        uint64
}
