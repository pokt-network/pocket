package telemetry

import (
	"log"
	"net/http"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ modules.TelemetryModule = &PromModule{}

// Prometheus struct
type PromModule struct {
	bus modules.Bus

	address  string
	endpoint string

	counters map[string]prometheus.Counter
	gauges   map[string]prometheus.Gauge
}

func CreatePromModule(cfg *config.Config) (*PromModule, error) {
	return &PromModule{
		counters: map[string]prometheus.Counter{},
		address:  cfg.Telemetry.Address,
		endpoint: cfg.Telemetry.Endpoint,
	}, nil
}

func (m *PromModule) Start() error {
	log.Println("Started the metrics exporter...")
	http.Handle(m.endpoint, promhttp.Handler())
	go http.ListenAndServe(m.address, nil)
	log.Println("Started OK")
	return nil
}

func (m *PromModule) Stop() error {
	return nil
}

func (m *PromModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *PromModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (p *PromModule) RegisterCounter(name string, description string) {
	if _, exists := p.counters[name]; exists {
		return
	}

	p.counters[name] = promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: description,
	})
}

func (p *PromModule) IncCounter(name string) {
	if _, exists := p.counters[name]; !exists {
		return
	}

	p.counters[name].Inc()
}

func (p *PromModule) RegisterGauge(name string, description string) {
	if _, exists := p.gauges[name]; exists {
		return
	}

	p.gauges[name] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: description,
	})
}

// Set sets the Gauge to an arbitrary value.
func (p *PromModule) SetGauge(name string, value float64) {
	if _, exists := p.gauges[name]; !exists {
		return
	}

	p.gauges[name].Set(value)
}

// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
// values.
func (p *PromModule) IncGauge(name string) {
	if _, exists := p.gauges[name]; !exists {
		return
	}

	p.gauges[name].Inc()
}

// IncGauge(name string)

// // Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
// // values.
func (p *PromModule) DecGauge(name string) {
	if _, exists := p.gauges[name]; !exists {
		return
	}

	p.gauges[name].Dec()
}

// // Add adds the given value to the Gauge. (The value can be negative,
// // resulting in a decrease of the Gauge.)
func (p *PromModule) AddToGauge(name string, value float64) {
	if _, exists := p.gauges[name]; !exists {
		return
	}

	p.gauges[name].Add(value)
}

// // Sub subtracts the given value from the Gauge. (The value can be
// // negative, resulting in an increase of the Gauge.)
func (p *PromModule) SubFromGauge(name string, value float64) {
	if _, exists := p.gauges[name]; !exists {
		return
	}

	p.gauges[name].Sub(value)
}
