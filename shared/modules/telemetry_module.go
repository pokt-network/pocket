package modules

type TelemetryModule interface {
	Module

	// metrics
	RegisterCounterMetric(name string, description string)
	IncrementCounterMetric(name string) error
}
