package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/telemetry_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import "github.com/prometheus/client_golang/prometheus"

const TelemetryModuleName = "telemetry"

type TelemetryModule interface {
	Module
	ConfigurableModule

	GetTimeSeriesAgent() TimeSeriesAgent
	GetEventMetricsAgent() EventMetricsAgent
}

// IMPROVE: Determine if the register function could (should?) return an error.

// Interface for the time series agent (prometheus)
type TimeSeriesAgent interface {
	/*** Counters ***/

	// Registers a counter by name
	CounterRegister(name string, description string)

	// Increments the counter
	CounterIncrement(name string) // DISCUSS(team): Should this return an error if the counter does not exist?

	/*** Gauges ***/

	// Register a gauge by name
	GaugeRegister(name string, description string)

	// Sets the Gauge to an arbitrary value
	GaugeSet(name string, value float64) (prometheus.Gauge, error)

	// Increments the Gauge by 1. Use Add to increment it by arbitrary values.
	GaugeIncrement(name string) (prometheus.Gauge, error)

	// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary values.
	GaugeDecrement(name string) (prometheus.Gauge, error)

	// Adds the given value to the Gauge. A negative value results in a decrease of the Gauge.
	GaugeAdd(name string, value float64) (prometheus.Gauge, error)

	// Subtracts the given value from the Gauge. A negative value results in a increase of the Gauge.
	GaugeSub(name string, value float64) (prometheus.Gauge, error)

	/*** Gauge Vectors ***/

	// Registers a gauge vector by name and provide labels
	GaugeVecRegister(namespace, module, name, description string, labels []string)

	// Retrieves a gauge vector by name
	GetGaugeVec(name string) (prometheus.GaugeVec, error)
}

// Interface for the event metrics agent
// IMPROVE: This relies on logging at the moment and can be improved in the future
type EventMetricsAgent interface {
	EmitEvent(namespace, event_name string, labels ...any)
}
