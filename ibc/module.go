package ibc

import (
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.IBCModule = &ibcModule{}

type ibcModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	// Only if a node is a validator on the network can it be an IBC host
	enabled bool

	// Only a single host is allowed at a time
	host *host
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(ibcModule).Create(bus, options...)
}

func (m *ibcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	ibcCfg := runtimeMgr.GetConfig().IBC

	if runtimeMgr.GetConfig().Validator.Enabled && ibcCfg.Enabled {
		m.enabled = true
	}

	return m, nil
}

func (m *ibcModule) Start() error {
	if !m.enabled {
		return nil
	}
	m.logger.Info().Msg("🪐 starting IBC module 🪐")
	m.logger.Info().Msg("🛰️ creating IBC host 🛰️")
	_, err := m.newHost()
	if err != nil {
		m.logger.Error().Err(err).Msg("❌ failed to create IBC host")
		return err
	}
	return nil
}

func (m *ibcModule) Stop() error {
	return nil
}

func (m *ibcModule) GetHost() modules.IBCHost {
	return m.host
}

func (m *ibcModule) GetModuleName() string {
	return modules.IBCModuleName
}

// newHost returns a new IBC host instance if one is not already created
func (m *ibcModule) newHost() (modules.IBCHost, error) {
	if m.host != nil {
		return nil, coreTypes.ErrHostAlreadyExists()
	}

	host := &host{
		logger: m.logger,
	}

	m.host = host

	return host, nil
}
