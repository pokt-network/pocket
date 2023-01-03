package telemetry

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	_ modules.Module             = &PrometheusTelemetryModule{}
	_ modules.ConfigurableModule = &PrometheusTelemetryModule{}
	_ modules.TelemetryModule    = &PrometheusTelemetryModule{}
	_ modules.EventMetricsAgent  = &PrometheusTelemetryModule{}
	_ modules.TimeSeriesAgent    = &PrometheusTelemetryModule{}
)

// DISCUSS(team): Should the warning logs in this module be handled differently?

type PrometheusTelemetryModule struct {
	bus    modules.Bus
	config modules.TelemetryConfig

	logger modules.Logger

	counters     map[string]prometheus.Counter
	gauges       map[string]prometheus.Gauge
	gaugeVectors map[string]prometheus.GaugeVec
}

const (
	prometheusModuleName = "prometheus"
)

func CreatePrometheusTelemetryModule(runtime modules.RuntimeMgr) (modules.Module, error) {
	var m PrometheusTelemetryModule
	return m.Create(runtime)
}

func (m *PrometheusTelemetryModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	telemetryCfg := cfg.GetTelemetryConfig()

	return &PrometheusTelemetryModule{
		config:       telemetryCfg,
		counters:     map[string]prometheus.Counter{},
		gauges:       map[string]prometheus.Gauge{},
		gaugeVectors: map[string]prometheus.GaugeVec{},
	}, nil
}

func (m *PrometheusTelemetryModule) Start() error {
	uri := m.config.GetAddress() + m.config.GetEndpoint()

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	m.logger.Info().Str("address", uri).Msg("Starting Prometheus metrics exporter...")

	http.Handle(m.config.GetEndpoint(), promhttp.Handler())
	go http.ListenAndServe(m.config.GetAddress(), nil)

	m.logger.Info().Str("address", uri).Msg("Prometheus metrics exporter started")

	return nil
}

func (m *PrometheusTelemetryModule) Stop() error {
	return nil
}

func (m *PrometheusTelemetryModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *PrometheusTelemetryModule) GetModuleName() string {
	return prometheusModuleName
}

func (m *PrometheusTelemetryModule) GetBus() modules.Bus {
	if m.bus == nil {
		m.logger.Fatal().Msg("PocketBus is not initialized")
	}
	return m.bus
}

func (*PrometheusTelemetryModule) ValidateConfig(cfg modules.Config) error {
	// TODO (#334): implement this
	return nil
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
	logArgs := map[string]interface{}{
		"level":     "EVENT",
		"namespace": namespace,
		"event":     event,
	}

	// iterate over the labels and appens them in pairs to the logArgs
	for i := 0; i < len(labels); i += 2 {
		logArgs[fmt.Sprintf("%v", labels[i])] = labels[i+1]
	}

	m.logger.Info().Fields(logArgs).Msg("Event emitted")
}

func (m *PrometheusTelemetryModule) GetTimeSeriesAgent() modules.TimeSeriesAgent {
	return modules.TimeSeriesAgent(m)
}

func (p *PrometheusTelemetryModule) CounterRegister(name string, description string) {
	if _, exists := p.counters[name]; exists {
		p.logger.Warn().Str("counter", name).Msg("Trying to register and already registered counter")
		return
	}

	p.counters[name] = promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: description,
	})
}

func (p *PrometheusTelemetryModule) CounterIncrement(name string) {
	if _, exists := p.counters[name]; !exists {
		p.logger.Warn().Str("counter", name).Msg("Trying to increment a non-existent counter")
		return
	}

	p.counters[name].Inc()
}

func (p *PrometheusTelemetryModule) GaugeRegister(name string, description string) {
	if _, exists := p.gauges[name]; exists {
		p.logger.Warn().Str("gauge", name).Msg("Trying to register and already registered gauge")
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
		p.logger.Warn().Str("gauge vector", name).Msg("Trying to register and already registered gauge vector")
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
