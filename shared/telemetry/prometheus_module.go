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
	_ modules.TelemetryModule   = &PrometheusTelemetryModule{}
	_ modules.EventMetricsAgent = &PrometheusTelemetryModule{}
	_ modules.TimeSeriesAgent   = &PrometheusTelemetryModule{}
)

// DISCUSS(team): Should the warning logs in this module be handled differently?

type PrometheusTelemetryModule struct {
	bus modules.Bus

	address  string
	endpoint string

	counters     map[string]prometheus.Counter
	gauges       map[string]prometheus.Gauge
	gaugeVectors map[string]prometheus.GaugeVec
}

func CreatePrometheusTelemetryModule(cfg *config.Config) (*PrometheusTelemetryModule, error) {
	return &PrometheusTelemetryModule{
		counters:     map[string]prometheus.Counter{},
		gauges:       map[string]prometheus.Gauge{},
		gaugeVectors: map[string]prometheus.GaugeVec{},

		address:  cfg.Telemetry.Address,
		endpoint: cfg.Telemetry.Endpoint,
	}, nil
}

func (m *PrometheusTelemetryModule) Start() error {
	log.Printf("\nPrometheus metrics exporter: Starting at %s/%s...\n", m.address, m.endpoint)

	http.Handle(m.endpoint, promhttp.Handler())
	go http.ListenAndServe(m.address, nil)

	log.Println("Prometheus metrics exporter started: OK")

	return nil
}

func (m *PrometheusTelemetryModule) Stop() error {
	return nil
}

func (m *PrometheusTelemetryModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *PrometheusTelemetryModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

// EventMetricsAgent interface implementation
func (m *PrometheusTelemetryModule) GetEventMetricsAgent() modules.EventMetricsAgent {
	return modules.EventMetricsAgent(m)
}

// At the moment, we are using loki to push log lines per event emission, and
// then we aggregate these log lines (count, avg, etc) in Grafana.
// Reliance on logs for event metrics is a temporary solution, and
// will be removed in the future in favor of more thorough event metrics tooling.
// TECHDEBT(team): Deprecate using logs for event metrics for a more sophisticated and durable solution
func (m *PrometheusTelemetryModule) EmitEvent(namespace, event string, labels ...any) {
	logArgs := append([]interface{}{"[EVENT]", namespace, event}, labels...)
	log.Println(logArgs...)
}

func (m *PrometheusTelemetryModule) GetTimeSeriesAgent() modules.TimeSeriesAgent {
	return modules.TimeSeriesAgent(m)
}

func (p *PrometheusTelemetryModule) CounterRegister(name string, description string) {
	if _, exists := p.counters[name]; exists {
		log.Printf("[WARNING] Trying to register and already registered counter: %s\n", name)
		return
	}

	p.counters[name] = promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: description,
	})
}

func (p *PrometheusTelemetryModule) CounterIncrement(name string) {
	if _, exists := p.counters[name]; !exists {
		log.Printf("[WARNING] Trying to increment a non-existent counter: %s\n", name)
		return
	}

	p.counters[name].Inc()
}

func (p *PrometheusTelemetryModule) GaugeRegister(name string, description string) {
	if _, exists := p.gauges[name]; exists {
		log.Printf("[WARNING] Trying to register and already registered gauge: %s\n", name)
		return
	}

	p.gauges[name] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: description,
	})
}

// Sets the Gauge to an arbitrary value.
func (p *PrometheusTelemetryModule) GaugeSet(name string, value float64) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistentMetricErr("gauge", name, "set")
	}

	gg := p.gauges[name]
	gg.Set(value)

	return gg, nil
}

// Increments the Gauge by 1. Use Add to increment it by arbitrary values.
func (p *PrometheusTelemetryModule) GaugeIncrement(name string) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistentMetricErr("gauge", name, "increment")
	}

	gg := p.gauges[name]
	gg.Inc()

	return gg, nil
}

func (p *PrometheusTelemetryModule) GaugeDecrement(name string) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistentMetricErr("gauge", name, "decrement")
	}

	gg := p.gauges[name]
	gg.Dec()

	return gg, nil
}

func (p *PrometheusTelemetryModule) GaugeAdd(name string, value float64) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistentMetricErr("gauge", name, "add to")
	}

	gg := p.gauges[name]
	gg.Add(value)

	return gg, nil
}

func (p *PrometheusTelemetryModule) GaugeSub(name string, value float64) (prometheus.Gauge, error) {
	if _, exists := p.gauges[name]; !exists {
		return nil, NonExistentMetricErr("gauge", name, "substract from")
	}

	gg := p.gauges[name]
	gg.Sub(value)
	return gg, nil
}

func (p *PrometheusTelemetryModule) GaugeVecRegister(namespace, module, name, description string, labels []string) {
	if _, exists := p.counters[name]; exists {
		log.Printf("[WARNING] Trying to register and already registered vector gauge: %s\n", name)
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
	p.gaugeVectors[name] = *gg
}

func (p *PrometheusTelemetryModule) GetGaugeVec(name string) (prometheus.GaugeVec, error) {
	if gv, exists := p.gaugeVectors[name]; exists {
		return gv, NonExistentMetricErr("gauge vector", name, "get")
	}
	return prometheus.GaugeVec{}, nil
}
