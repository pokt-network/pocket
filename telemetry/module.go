package telemetry

import (
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.Module = &telemetryModule{}
var _ modules.TelemetryConfig = &TelemetryConfig{}

const (
	TelemetryModuleName = "telemetry"
)

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func Create(cfg modules.TelemetryConfig) (modules.TelemetryModule, error) {
	moduleCfg := cfg.(*TelemetryConfig)
	if moduleCfg.GetEnabled() {
		return CreatePrometheusTelemetryModule(moduleCfg)
	} else {
		return CreateNoopTelemetryModule(moduleCfg)
	}
}

type telemetryModule struct{}

func (t *telemetryModule) GetModuleName() string                                      { return TelemetryModuleName }
func (t *telemetryModule) InitGenesis(_ string) (genesis modules.IGenesis, err error) { return }
func (t *telemetryModule) SetBus(bus modules.Bus)                                     {}
func (t *telemetryModule) GetBus() modules.Bus                                        { return nil }
func (t *telemetryModule) Start() error                                               { return nil }
func (t *telemetryModule) Stop() error                                                { return nil }
