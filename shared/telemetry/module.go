package telemetry

import (
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
)

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func Create(cfg *config.Config) (modules.TelemetryModule, error) {
	if cfg.EnableTelemetry {
		return CreatePrometheusTelemetryModule(cfg)
	} else {
		return CreateNoopTelemetryModule(cfg)
	}
}
