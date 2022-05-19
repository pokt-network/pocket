package modules

type TelemetryModule interface {
	Module

	// Counters

	// Register a counter by name
	RegisterCounter(name string, description string)
	// Increment the counter
	IncCounter(name string)

	// Gauges

	// Register a gauge by name
	RegisterGauge(name string, description string)
	// Set sets the Gauge to an arbitrary value.
	SetGauge(name string, value float64)
	// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
	// values.
	IncGauge(name string)
	// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
	// values.
	DecGauge(name string)
	// Add adds the given value to the Gauge. (The value can be negative,
	// resulting in a decrease of the Gauge.)
	AddToGauge(name string, value float64)
	// Sub subtracts the given value from the Gauge. (The value can be
	// negative, resulting in an increase of the Gauge.)
	SubFromGauge(name string, value float64)
}
