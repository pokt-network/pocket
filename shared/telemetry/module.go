package telemetry

import (
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
)

func Create(cfg *config.Config) (modules.TelemetryModule, error) {
	// TODO(team): Add a switch statement and configuration variable when support for other telemetry modules is added.
	if cfg.EnableTelemetry {
		return CreatePrometheusTelemetryModule(cfg)
	} else {
		return CreateNoopTelemetryModule(cfg)
	}
}
