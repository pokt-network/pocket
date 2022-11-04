package runtime

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	typesLogger "github.com/pokt-network/pocket/logger"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	typesPers "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
	typesTelemetry "github.com/pokt-network/pocket/telemetry"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

var _ modules.Config = &runtimeConfig{}

type runtimeConfig struct {
	Base        *BaseConfig                     `json:"base"`
	Consensus   *typesCons.ConsensusConfig      `json:"consensus"`
	Utility     *typesUtil.UtilityConfig        `json:"utility"`
	Persistence *typesPers.PersistenceConfig    `json:"persistence"`
	P2P         *typesP2P.P2PConfig             `json:"p2p"`
	Telemetry   *typesTelemetry.TelemetryConfig `json:"telemetry"`
	Logger      *typesLogger.LoggerConfig       `json:"logger"`
}

func NewConfig(base *BaseConfig, otherConfigs ...func(modules.Config)) *runtimeConfig {
	rc := &runtimeConfig{
		Base: base,
	}
	for _, oc := range otherConfigs {
		oc(rc)
	}
	return rc
}

func WithConsensusConfig(consensusConfig modules.ConsensusConfig) func(modules.Config) {
	return func(rc modules.Config) {
		rc.(*runtimeConfig).Consensus = consensusConfig.(*typesCons.ConsensusConfig)
	}
}

func WithUtilityConfig(utilityConfig modules.UtilityConfig) func(modules.Config) {
	return func(rc modules.Config) {
		rc.(*runtimeConfig).Utility = utilityConfig.(*typesUtil.UtilityConfig)
	}
}

func WithPersistenceConfig(persistenceConfig modules.PersistenceConfig) func(modules.Config) {
	return func(rc modules.Config) {
		rc.(*runtimeConfig).Persistence = persistenceConfig.(*typesPers.PersistenceConfig)
	}
}

func WithP2PConfig(p2pConfig modules.P2PConfig) func(modules.Config) {
	return func(rc modules.Config) {
		rc.(*runtimeConfig).P2P = p2pConfig.(*typesP2P.P2PConfig)
	}
}

func WithTelemetryConfig(telemetryConfig modules.TelemetryConfig) func(modules.Config) {
	return func(rc modules.Config) {
		rc.(*runtimeConfig).Telemetry = telemetryConfig.(*typesTelemetry.TelemetryConfig)
	}
}

func (c *runtimeConfig) GetBaseConfig() modules.BaseConfig {
	return c.Base
}

func (c *runtimeConfig) GetConsensusConfig() modules.ConsensusConfig {
	return c.Consensus
}

func (c *runtimeConfig) GetUtilityConfig() modules.UtilityConfig {
	return c.Utility
}

func (c *runtimeConfig) GetPersistenceConfig() modules.PersistenceConfig {
	return c.Persistence
}

func (c *runtimeConfig) GetP2PConfig() modules.P2PConfig {
	return c.P2P
}

func (c *runtimeConfig) GetTelemetryConfig() modules.TelemetryConfig {
	return c.Telemetry
}

func (c *runtimeConfig) GetLoggerConfig() modules.LoggerConfig {
	return c.Logger
}
