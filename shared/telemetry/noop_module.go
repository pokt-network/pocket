package telemetry

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/prometheus/client_golang/prometheus"
)

var _ modules.TelemetryModule = &PromModule{}

// Prometheus struct
type NoopModule struct {
	bus modules.Bus

	address  string
	endpoint string

	counters map[string]prometheus.Counter
	gauges   map[string]prometheus.Gauge
}

func NOOP() {
	log.Printf("\n[telemetry=noop]\n")
}

func CreateNoopModule(cfg *config.Config) (*NoopModule, error) {
	return &NoopModule{}, nil
}

func (m *NoopModule) Start() error {
	NOOP()
	return nil
}

func (m *NoopModule) Stop() error {
	NOOP()
	return nil
}

func (m *NoopModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *NoopModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (p *NoopModule) RegisterCounter(name string, description string) { NOOP() }

func (p *NoopModule) IncCounter(name string) { NOOP() }

func (p *NoopModule) RegisterGauge(name string, description string) {
	NOOP()
}

// Set sets the Gauge to an arbitrary value.
func (p *NoopModule) SetGauge(name string, value float64) prometheus.Gauge { NOOP(); return nil }

// IncGauge(name string)
// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
// values.
func (p *NoopModule) IncGauge(name string) prometheus.Gauge {
	NOOP()
	return nil
}

// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
// values.
func (p *NoopModule) DecGauge(name string) prometheus.Gauge { NOOP(); return nil }

// Add adds the given value to the Gauge. (The value can be negative,
// resulting in a decrease of the Gauge.)
func (p *NoopModule) AddToGauge(name string, value float64) prometheus.Gauge { NOOP(); return nil }

// Sub subtracts the given value from the Gauge. (The value can be
// negative, resulting in an increase of the Gauge.)
func (p *NoopModule) SubFromGauge(name string, value float64) prometheus.Gauge { NOOP(); return nil }

func (p *NoopModule) GetGaugeVec(name string) *prometheus.GaugeVec { NOOP(); return nil }
func (p *NoopModule) RegisterGaugeVector(namespace, module, name, description string, labels []string) {
	NOOP()
}
