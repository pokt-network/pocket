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

func NOOP(args ...interface{}) {
	log.Printf("\n[telemetry=noop][%s]\n", args)
}

func CreateNoopTelemetryModule(bus modules.Bus) (modules.Module, error) {
	var m NoopTelemetryModule
	return m.Create(bus)
}

func (*NoopTelemetryModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &NoopTelemetryModule{}
	bus.RegisterModule(m)
	return m, nil
}

func (*NoopTelemetryModule) Start() error {
	NOOP("Start")
	return nil
}

func (*NoopTelemetryModule) Stop() error {
	NOOP("Stop")
	return nil
}

func (*NoopTelemetryModule) GetModuleName() string {
	return fmt.Sprintf("%s_noOP", modules.TelemetryModuleName)
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
	return modules.EventMetricsAgent(m)
}

func (*NoopTelemetryModule) EmitEvent(namespace, event_name string, labels ...any) {
	NOOP("EmitEvent", "namespace", namespace, "event_name", event_name, "labels", labels)
}

func (m *NoopTelemetryModule) GetTimeSeriesAgent() modules.TimeSeriesAgent {
	return modules.TimeSeriesAgent(m)
}

func (*NoopTelemetryModule) CounterRegister(name string, description string) {
	NOOP("CounterRegister", "name", name, "description", description)
}

func (*NoopTelemetryModule) CounterIncrement(name string) {
	NOOP("CounterIncrement", "name", name)
}

func (*NoopTelemetryModule) GaugeRegister(name string, description string) {
	NOOP("GaugeRegister", "name", name, "description", description)
}

func (*NoopTelemetryModule) GaugeSet(name string, value float64) (prometheus.Gauge, error) {
	NOOP("GaugeSet", "name", name, "value", value)
	return nil, nil
}

func (*NoopTelemetryModule) GaugeIncrement(name string) (prometheus.Gauge, error) {
	NOOP("GaugeIncrement", "name", name)
	return nil, nil
}

func (*NoopTelemetryModule) GaugeDecrement(name string) (prometheus.Gauge, error) {
	NOOP("GaugeDecrement", "name", name)
	return nil, nil
}

func (*NoopTelemetryModule) GaugeAdd(name string, value float64) (prometheus.Gauge, error) {
	NOOP("GaugeAdd", "name", name, "value", value)
	return nil, nil
}

func (*NoopTelemetryModule) GaugeSub(name string, value float64) (prometheus.Gauge, error) {
	NOOP("GaugeSub", "name", name, "value", value)
	return nil, nil
}

func (*NoopTelemetryModule) GetGaugeVec(name string) (prometheus.GaugeVec, error) {
	NOOP("GetGaugeVec", "name", name)
	return prometheus.GaugeVec{}, nil
}

func (*NoopTelemetryModule) GaugeVecRegister(namespace, module, name, description string, labels []string) {
	NOOP("GaugeVecRegister", "namespace", namespace, "module", module, "name", name, "description", description, "labels", labels)
}
