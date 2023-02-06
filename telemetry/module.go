package telemetry

import (
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_                   modules.Module = &telemetryModule{}
	ImplementationNames                = []string{
		new(PrometheusTelemetryModule).GetModuleName(),
		new(NoopTelemetryModule).GetModuleName(),
	}
)

type telemetryModule struct{}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(telemetryModule).Create(bus)
}

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func (*telemetryModule) Create(bus modules.Bus) (modules.Module, error) {
	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()

	telemetryCfg := cfg.Telemetry

	if telemetryCfg.Enabled {
		return CreatePrometheusTelemetryModule(bus)
	} else {
		return CreateNoopTelemetryModule(bus)
	}
}

func (t *telemetryModule) GetModuleName() string  { return modules.TelemetryModuleName }
func (t *telemetryModule) SetBus(bus modules.Bus) {}
func (t *telemetryModule) GetBus() modules.Bus    { return nil }
func (t *telemetryModule) Start() error           { return nil }
func (t *telemetryModule) Stop() error            { return nil }
