package modules

import (
	"github.com/pokt-network/pocket/shared/logging"
	"github.com/prometheus/client_golang/prometheus"
)

type TelemetryModule interface {
	Module

	GetTimeSeriesAgent() TimeSeriesAgent
	GetEventMetricsAgent() EventMetricsAgent

	LoggerGet(namespace logging.Namespace) logging.Logger
	LoggerRegister(namespace logging.Namespace, level logging.LogLevel) error
}

// Interface for the time series agent (prometheus)
type TimeSeriesAgent interface {
	/*** Counters ***/

	// Registers a counter by name
	CounterRegister(name string, description string)

	// Increments the counter
	CounterIncrement(name string)

	/*** Gauges ***/

	// Register a gauge by name
	GaugeRegister(name string, description string)

	// Sets the Gauge to an arbitrary value.
	GaugeSet(name string, value float64) (prometheus.Gauge, error)

	// Increments the Gauge by 1. Use Add to increment it by arbitrary values.
	GaugeIncrement(name string) (prometheus.Gauge, error)

	// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary values.
	GaugeDecrement(name string) (prometheus.Gauge, error)

	// Adds the given value to the Gauge. A negative value results in a decrease of the Gauge.
	GaugeAdd(name string, value float64) (prometheus.Gauge, error)

	// Subtracts the given value from the Gauge. (The value can be negative, resulting in an increase of the Gauge.)
	GaugeSub(name string, value float64) (prometheus.Gauge, error)

	/*** Gauge Vectors ***/

	// Registers a gauge vector by name and provide labels
	GaugeVecRegister(namespace, module, name, description string, labels []string)

	// Retrieves a gauge vector by name
	GetGaugeVec(name string) (prometheus.GaugeVec, error)
}

// Interface for the event metrics agent (relies on logging ftm)
type EventMetricsAgent interface {
	EmitEvent(namespace, event_name string, labels ...any)
}
