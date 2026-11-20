package gopsutil

import (
	"errors"
	"math"
	"testing"

	"github.com/shirou/gopsutil/v4/sensors"
)

func TestValidTemperaturesKeepsPartialReadingsAndSorts(t *testing.T) {
	readings, err := validTemperatures([]sensors.TemperatureStat{
		{SensorKey: "Drive", Temperature: 41},
		{SensorKey: "Broken", Temperature: math.NaN()},
		{SensorKey: "CPU", Temperature: 63.4},
		{SensorKey: "Board", Temperature: 41},
	}, errors.New("a sensor failed"))
	if err != nil || len(readings) != 3 {
		t.Fatalf("expected three valid partial readings, got %+v, %v", readings, err)
	}
	for i, name := range []string{"CPU", "Board", "Drive"} {
		if readings[i].Name != name {
			t.Errorf("reading %d: got %q, want %q", i, readings[i].Name, name)
		}
	}
}

func TestValidTemperaturesDistinguishesNoSensorsFromFailure(t *testing.T) {
	if readings, err := validTemperatures(nil, nil); len(readings) != 0 || err != nil {
		t.Fatalf("expected empty successful reading, got %+v, %v", readings, err)
	}
	failure := errors.New("sensor poll failed")
	if readings, err := validTemperatures(nil, failure); len(readings) != 0 || !errors.Is(err, failure) {
		t.Fatalf("expected collection error, got %+v, %v", readings, err)
	}
}
