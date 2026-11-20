package telemetry

import "time"

type HostInfo struct {
	Hostname        string
	BootTime        time.Time
	Procs           uint64
	OS              string
	Platform        string
	PlatformFamily  string
	PlatformVersion string
	KernelVersion   string
	KernelArch      string
}
