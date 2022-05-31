package telemetry

import (
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
)

func Create(cfg *config.Config) (modules.TelemetryModule, error) {
	if cfg.UseTelemetry {
		return CreatePromModule(cfg)
	} else {
		return CreateNoopModule(cfg)
	}
}
