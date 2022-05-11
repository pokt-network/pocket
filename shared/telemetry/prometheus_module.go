package telemetry

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ modules.TelemetryModule = &PromModule{}

type PromModule struct {
	bus modules.Bus

	address  string
	endpoint string

	counters map[string]prometheus.Counter
}

func Create(cfg *config.Config) (*PromModule, error) {
	return &PromModule{
		counters: map[string]prometheus.Counter{},
		address:  cfg.Telemetry.Address,
		endpoint: cfg.Telemetry.Endpoint,
	}, nil
}

func (m *PromModule) Start() error {
	http.Handle(m.endpoint, promhttp.Handler())
	go http.ListenAndServe(m.address, nil)
	log.Println("Started the metrics exporter...")
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

func (p *PromModule) RegisterCounterMetric(name string, description string) {
	if _, exists := p.counters[name]; exists {
		return
	}

	p.counters[name] = promauto.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: description,
	})
}

func (p *PromModule) IncrementCounterMetric(name string) error {
	if _, exists := p.counters[name]; !exists {
		return fmt.Errorf("Prometheus Instrument: Trying to increment a non-tracked counter")
	}

	p.counters[name].Inc()
	return nil
}
