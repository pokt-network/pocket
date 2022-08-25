package telemetry

import "fmt"

var (
	NonExistentMetricErr = func(metricType, name, action string) error {
		return fmt.Errorf("tried to %s a non-existent %s: %s", action, name, metricType)
	}
)
