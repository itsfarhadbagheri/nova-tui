package gopsutil

import (
	"context"
	"math"
	"slices"
	"strings"

	"github.com/itsfarhadbagheri/nova-tui/internal/telemetry"
	"github.com/shirou/gopsutil/v4/sensors"
)

type TemperatureCollector struct{}

func NewTemperatureCollector() *TemperatureCollector {
	return &TemperatureCollector{}
}

func (c *TemperatureCollector) Collect(ctx context.Context) ([]telemetry.Temperature, error) {
	readings, err := sensors.TemperaturesWithContext(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return validTemperatures(readings, err)
}

func validTemperatures(readings []sensors.TemperatureStat, err error) ([]telemetry.Temperature, error) {
	var result []telemetry.Temperature
	for _, reading := range readings {
		value := reading.Temperature
		if math.IsNaN(value) || math.IsInf(value, 0) || value < -273.15 {
			continue
		}
		name := strings.TrimSpace(reading.SensorKey)
		if name == "" {
			name = "Unnamed sensor"
		}
		result = append(result, telemetry.Temperature{Name: name, Celsius: value})
	}
	if len(result) == 0 {
		return nil, err
	}
	slices.SortFunc(result, func(a, b telemetry.Temperature) int {
		if a.Celsius > b.Celsius {
			return -1
		}
		if a.Celsius < b.Celsius {
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return result, nil
}
