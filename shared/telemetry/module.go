package telemetry

import (
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func Create(cfg *genesis.Config, _ *genesis.GenesisState) (modules.TelemetryModule, error) {
	if cfg.Telemetry.Enabled {
		return CreatePrometheusTelemetryModule(cfg)
	} else {
		return CreateNoopTelemetryModule(cfg)
	}
}
