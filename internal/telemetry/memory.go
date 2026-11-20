package telemetry

type MemoryUsage struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	Free        uint64
	Active      uint64
}
