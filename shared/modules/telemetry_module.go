package modules

import "github.com/prometheus/client_golang/prometheus"

type (
	TelemetryModule interface {
		Module

		// Time series
		GetTimeSeriesAgent() TimeSeriesAgent
		GetEventMetricsAgent() EventMetricsAgent
	}

	// Interface for time series agent (prometheus)
	TimeSeriesAgent interface {
		// Counters

		// Register a counter by name
		RegisterCounter(name string, description string)
		// Increment the counter
		IncCounter(name string)

		// Gauges

		// Register a gauge by name
		RegisterGauge(name string, description string)
		// Set sets the Gauge to an arbitrary value.
		SetGauge(name string, value float64) prometheus.Gauge
		// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
		// values.
		IncGauge(name string) prometheus.Gauge
		// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
		// values.
		DecGauge(name string) prometheus.Gauge
		// Add adds the given value to the Gauge. (The value can be negative,
		// resulting in a decrease of the Gauge.)
		AddToGauge(name string, value float64) prometheus.Gauge
		// Sub subtracts the given value from the Gauge. (The value can be
		// negative, resulting in an increase of the Gauge.)
		SubFromGauge(name string, value float64) prometheus.Gauge

		// Gauge Vectors

		// Register a gauge vector by name and provide labels
		RegisterGaugeVector(namespace, module, name, description string, labels []string)
		// Retrieve a gauge vector by name
		GetGaugeVec(name string) *prometheus.GaugeVec
	}

	// Interface for the event metrics agent (relies on logging ftm)
	EventMetricsAgent interface {
		EmitEvent(...interface{})
	}
)
