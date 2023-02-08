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

type telemetryModule struct {
	modules.BaseIntegratableModule
	modules.BaseInterruptableModule
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(telemetryModule).Create(bus, options...)
}

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func (*telemetryModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()

	telemetryCfg := cfg.Telemetry

	if telemetryCfg.Enabled {
		return CreatePrometheusTelemetryModule(bus)
	} else {
		return CreateNoopTelemetryModule(bus)
	}
}

func (t *telemetryModule) GetModuleName() string { return modules.TelemetryModuleName }
