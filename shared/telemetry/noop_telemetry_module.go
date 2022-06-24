package telemetry

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ modules.TelemetryModule   = &NoopTelemetryModule{}
	_ modules.EventMetricsAgent = &NoopTelemetryModule{}
	_ modules.TimeSeriesAgent   = &NoopTelemetryModule{}
)

type NoopTelemetryModule struct {
	bus modules.Bus
}

func NOOP() {
	log.Printf("\n[telemetry=noop]\n")
}

func CreateNoopTelemetryModule(cfg *config.Config) (*NoopTelemetryModule, error) {
	return &NoopTelemetryModule{}, nil
}

func (m *NoopTelemetryModule) Start() error {
	NOOP()
	return nil
}

func (m *NoopTelemetryModule) Stop() error {
	NOOP()
	return nil
}

func (m *NoopTelemetryModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *NoopTelemetryModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *NoopTelemetryModule) GetEventMetricsAgent() modules.EventMetricsAgent {
	return m
}

func (m *NoopTelemetryModule) EmitEvent(namespace, event_name string, labels ...any) {

}

func (m *NoopTelemetryModule) GetTimeSeriesAgent() modules.TimeSeriesAgent {
	return m
}

func (p *NoopTelemetryModule) CounterRegister(name string, description string) { NOOP() }

func (p *NoopTelemetryModule) CounterIncrement(name string) { NOOP() }

func (p *NoopTelemetryModule) GaugeRegister(name string, description string) {
	NOOP()
}

// Set sets the Gauge to an arbitrary value.
func (p *NoopTelemetryModule) GaugeSet(name string, value float64) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

// Increments the Gauge by 1. Use Add to increment it by arbitrary values.
func (p *NoopTelemetryModule) GaugeIncrement(name string) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

// Decrements the Gauge by 1. Use Sub to decrement it by arbitrary
func (p *NoopTelemetryModule) GaugeDecrement(name string) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

// Add adds the given value to the Gauge. (The value can be negative, resulting in a decrease of the Gauge.)
func (p *NoopTelemetryModule) GaugeAdd(name string, value float64) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

// Subtracts the given value from the Gauge. (The value can be negative, resulting in an increase of the Gauge.)
func (p *NoopTelemetryModule) GaugeSub(name string, value float64) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

func (p *NoopTelemetryModule) GetGaugeVec(name string) (prometheus.GaugeVec, error) {
	NOOP()
	return *&prometheus.GaugeVec{}, nil
}

func (p *NoopTelemetryModule) GaugeVecRegister(namespace, module, name, description string, labels []string) {
	NOOP()
}
