package telemetry

type DiskUsage struct {
	Path        string
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}
