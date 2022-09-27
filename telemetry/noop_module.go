package telemetry

import (
	"fmt"
	"log"

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

const (
	NoOpModuleName = "noOP"
)

func NOOP() {
	log.Printf("\n[telemetry=noop]\n")
}

func CreateNoopTelemetryModule(runtime modules.Runtime) (modules.Module, error) {
	var m NoopTelemetryModule
	return m.Create(runtime)
}

func (m *NoopTelemetryModule) Create(runtime modules.Runtime) (modules.Module, error) {
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

func (m *NoopTelemetryModule) GetModuleName() string {
	return NoOpModuleName
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

func (*NoopTelemetryModule) ValidateConfig(cfg modules.Config) error {
	if _, ok := cfg.Telemetry.(*TelemetryConfig); !ok {
		return fmt.Errorf("cannot cast to TelemetryConfig")
	}
	return nil
}

func (m *NoopTelemetryModule) GetEventMetricsAgent() modules.EventMetricsAgent {
	return modules.EventMetricsAgent(m)
}

func (m *NoopTelemetryModule) EmitEvent(namespace, event_name string, labels ...any) {
	NOOP()
}

func (m *NoopTelemetryModule) GetTimeSeriesAgent() modules.TimeSeriesAgent {
	return modules.TimeSeriesAgent(m)
}

func (p *NoopTelemetryModule) CounterRegister(name string, description string) {
	NOOP()
}

func (p *NoopTelemetryModule) CounterIncrement(name string) {
	NOOP()
}

func (p *NoopTelemetryModule) GaugeRegister(name string, description string) {
	NOOP()
}

func (p *NoopTelemetryModule) GaugeSet(name string, value float64) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

func (p *NoopTelemetryModule) GaugeIncrement(name string) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

func (p *NoopTelemetryModule) GaugeDecrement(name string) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

func (p *NoopTelemetryModule) GaugeAdd(name string, value float64) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

func (p *NoopTelemetryModule) GaugeSub(name string, value float64) (prometheus.Gauge, error) {
	NOOP()
	return nil, nil
}

func (p *NoopTelemetryModule) GetGaugeVec(name string) (prometheus.GaugeVec, error) {
	NOOP()
	return prometheus.GaugeVec{}, nil
}

func (p *NoopTelemetryModule) GaugeVecRegister(namespace, module, name, description string, labels []string) {
	NOOP()
}
