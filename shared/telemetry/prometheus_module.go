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

var (
	_ modules.TelemetryModule   = &PromModule{}
	_ modules.EventMetricsAgent = &PromModule{}
	_ modules.TimeSeriesAgent   = &PromModule{}
)

type PromModule struct {
	bus modules.Bus

	address  string
	endpoint string

	counters     map[string]prometheus.Counter
	gauges       map[string]prometheus.Gauge
	gaugeVectors map[string]prometheus.GaugeVec
}

func CreatePromModule(cfg *config.Config) (*PromModule, error) {
	return &PromModule{
		counters:     map[string]prometheus.Counter{},
		gauges:       map[string]prometheus.Gauge{},
		gaugeVectors: map[string]prometheus.GaugeVec{},

		address:  cfg.Telemetry.Address,
		endpoint: cfg.Telemetry.Endpoint,
	}, nil
}

func (m *PromModule) Start() error {
	log.Printf("\nPrometheus metrics exporter: Starting at %s/%s...\n", m.address, m.endpoint)

	http.Handle(m.endpoint, promhttp.Handler())
	go http.ListenAndServe(m.address, nil)

	log.Println("Prometheus metrics exporter started: OK")

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

// Event Metrics functionality implementation
func (m *PromModule) GetEventMetricsAgent() modules.EventMetricsAgent {
	return m.(modules.EventMetricsAgent)
}

// At the moment, we are using loki to push log lines per event emission, and
// then we aggregate these log lines (count them, avg, etc) in grafana.
// Reliance on logs for event metrics was a temporary solution, and
// will be removed in the future in favor of more thorough event metrics tooling.
func (m *PromModule) EmitEvent(namespace, event string, labels ...any{}) {
	logArgs := append([]interface{}{"[EVENT]", namespace, event}, labels...)
	log.Println(logArgs...)
}

func (m *PromModule) GetTimeSeriesAgent() modules.TimeSeriesAgent {
	return m.(modules.TimeSeriesAgent)
}

func (p *PromModule) CounterRegister(name string, description string) {
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

func (p *PromModule) GaugeRegister(name string, description string) {
	if _, exists := p.gauges[name]; exists {
		return
	}

	p.gauges[name] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: description,
	})
}

// Sets the Gauge to an arbitrary value.
func (p *PromModule) GaugeSet(name string, value float64) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistantMetric("gauge", name, "set")
	}

	gg := p.gauges[name]
	gg.Set(value)

	return gg
}

// Increments the Gauge by 1. Use Add to increment it by arbitrary values.
func (p *PromModule) GaugeIncrement(name string) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistentErr("gauge", name, "increment")
	}

	gg := p.gauges[name]
	gg.Inc()

	return gg
}

// Decrements the Gauge by 1. Use Sub to decrement it by arbitrary values.
func (p *PromModule) GaugeDecrement(name string) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistantMetric("gauge", name, "decrement")
	}

	gg := p.gauges[name]
	gg.Dec()

	return gg
}

// Adds the given value to the Gauge. (The value can be negative, resulting in a decrease of the Gauge.)
func (p *PromModule) GaugeAdd(name string, value float64) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistantMetric("gauge", name, "add to")
	}

	gg := p.gauges[name]
	gg.Add(value)

	return gg
}

// Subtracts the given value from the Gauge. (The value can be negative, resulting in an increase of the Gauge.)
func (p *PromModule) GaugeSubstract(name string, value float64) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistantMetric("gauge", name, "substract from")
	}

	gg := p.gauges[name]
	gg.Sub(value)
	return gg
}

// Registers a gauge vector by name and provide labels
func (p *PromModule) GaugeVecRegister(namespace, module, name, description string, labels []string) {
	if _, exists := p.counters[name]; exists {
		return
	}

	gg := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: module,
			Name:      name,
			Help:      description,
		},
		labels,
	)
	p.gaugeVectors[name] = gg
}

// Retrieves a gauge vector by name
func (p *PromModule) GetGaugeVec(name string)( prometheus.GaugeVec, error) {
	if gv, exists := p.gaugeVectors[name]; exists {
		return gv, NonExistantMetric("gauge vector", name, "get")
	}
	return nil
}
