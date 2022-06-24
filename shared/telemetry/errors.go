package telemetry

import "fmt"

var (
	NonExistantMetricErr = func(mtype, name, action string) error {
		return fmt.Errorf("Tried to %s a non-existant %s: %s", action, name, mtype)
	}
)
