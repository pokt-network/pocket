package telemetry

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.TelemetryConfig = &TelemetryConfig{}

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func Create(config, genesis json.RawMessage) (modules.TelemetryModule, error) {
	cfg, err := InitConfig(config)
	if err != nil {
		return nil, err
	}
	if cfg.GetEnabled() {
		return CreatePrometheusTelemetryModule(cfg)
	} else {
		return CreateNoopTelemetryModule(cfg)
	}
}

func InitGenesis(data json.RawMessage) {
	// TODO (Team) add genesis state if necessary
}

func InitConfig(data json.RawMessage) (config *TelemetryConfig, err error) {
	config = new(TelemetryConfig)
	err = json.Unmarshal(data, config)
	return
}
