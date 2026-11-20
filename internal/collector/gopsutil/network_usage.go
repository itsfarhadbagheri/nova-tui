package gopsutil

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	gopsnet "github.com/shirou/gopsutil/v4/net"
)

type NetworkUsageCollector struct {
	mu       sync.RWMutex
	loopback map[string]bool
	known    map[string]struct{}
}

func NewNetworkUsageCollector() *NetworkUsageCollector {
	collector := &NetworkUsageCollector{}
	_ = collector.refreshInterfaces()
	return collector
}

func (c *NetworkUsageCollector) Collect(ctx context.Context) (telemetry.NetworkCounters, error) {
	if err := ctx.Err(); err != nil {
		return telemetry.NetworkCounters{}, err
	}
	counters, err := gopsnet.IOCountersWithContext(ctx, true)
	if err != nil {
		return telemetry.NetworkCounters{}, fmt.Errorf("read per-interface network counters: %w", err)
	}

	if c.hasUnknownInterfaces(counters) {
		if err := c.refreshInterfaces(); err != nil {
			return telemetry.NetworkCounters{}, fmt.Errorf("refresh network interface metadata: %w", err)
		}
	}

	result := telemetry.NetworkCounters{Interfaces: make([]telemetry.InterfaceCounters, 0, len(counters))}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, counter := range counters {
		loopback, known := c.loopback[counter.Name]
		if !known || loopback {
			continue
		}
		result.Interfaces = append(result.Interfaces, telemetry.InterfaceCounters{
			Name:      counter.Name,
			BytesSent: counter.BytesSent,
			BytesRecv: counter.BytesRecv,
		})
	}
	return result, nil
}

func (c *NetworkUsageCollector) hasUnknownInterfaces(counters []gopsnet.IOCountersStat) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, counter := range counters {
		if _, ok := c.known[counter.Name]; !ok {
			return true
		}
	}
	return false
}

func (c *NetworkUsageCollector) refreshInterfaces() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	known := make(map[string]struct{}, len(interfaces))
	loopback := make(map[string]bool, len(interfaces))
	for _, iface := range interfaces {
		known[iface.Name] = struct{}{}
		loopback[iface.Name] = iface.Flags&net.FlagLoopback != 0
	}
	c.mu.Lock()
	c.known = known
	c.loopback = loopback
	c.mu.Unlock()
	return nil
}
