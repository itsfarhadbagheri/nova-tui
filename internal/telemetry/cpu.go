package telemetry

type CPUInfo struct {
	ModelName        string
	PhysicalCores    int
	HasPhysicalCores bool
}

type CPUUsage struct {
	Usage float64
}
