package telemetry

import (
	"encoding/json"

	"github.com/pokt-network/pocket/shared/modules"
	typesTelemetry "github.com/pokt-network/pocket/telemetry/types"
)

var _ modules.TelemetryConfig = &typesTelemetry.TelemetryConfig{}

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

func InitConfig(data json.RawMessage) (config *typesTelemetry.TelemetryConfig, err error) {
	config = new(typesTelemetry.TelemetryConfig)
	err = json.Unmarshal(data, config)
	return
}
