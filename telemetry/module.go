package telemetry

import (
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.Module          = &telemetryModule{}
	_ modules.TelemetryConfig = &TelemetryConfig{}
)

const (
	TelemetryModuleName = "telemetry"
)

func Create(runtime modules.Runtime) (modules.Module, error) {
	return new(telemetryModule).Create(runtime)
}

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func (*telemetryModule) Create(runtime modules.Runtime) (modules.Module, error) {
	cfg := runtime.GetConfig()

	telemetryCfg := cfg.Telemetry.(*TelemetryConfig)

	if telemetryCfg.GetEnabled() {
		return CreatePrometheusTelemetryModule(runtime)
	} else {
		return CreateNoopTelemetryModule(runtime)
	}
}

type telemetryModule struct{}

func (t *telemetryModule) GetModuleName() string                                      { return TelemetryModuleName }
func (t *telemetryModule) InitGenesis(_ string) (genesis modules.IGenesis, err error) { return }
func (t *telemetryModule) SetBus(bus modules.Bus)                                     {}
func (t *telemetryModule) GetBus() modules.Bus                                        { return nil }
func (t *telemetryModule) Start() error                                               { return nil }
func (t *telemetryModule) Stop() error                                                { return nil }
