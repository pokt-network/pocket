package telemetry

import "fmt"

var (
	NonExistentMetricErr = func(metricType, name, action string) error {
		return fmt.Errorf("Tried to %s a non-existant %s: %s", action, name, metricType)
	}
)
