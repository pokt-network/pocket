package ibc

import (
	"github.com/pokt-network/pocket/ibc/stores"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.IBCModule = &ibcModule{}

type ibcModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	hostEnabled bool
	storesDir   string

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
	m.hostEnabled = ibcCfg.HostEnabled
	m.storesDir = ibcCfg.StoresDir

	return m, nil
}

func (m *ibcModule) Start() error {
	m.logger.Info().Msg("🪐 starting IBC module 🪐")
	if m.hostEnabled {
		m.logger.Info().Msg("🛰️ creating IBC host 🛰️")
		_, err := m.NewHost()
		if err != nil {
			m.logger.Error().Err(err).Msg("❌ failed to create IBC host")
			return err
		}
	}
	return nil
}

func (m *ibcModule) Stop() error {
	if m.host != nil {
		m.logger.Info().Msg("🚨 closing IBC host stores 🚨")
		return m.host.GetStoreManager().CloseAllStores()
	}
	return nil
}

func (m *ibcModule) GetModuleName() string {
	return modules.IBCModuleName
}

// NewHost returns a new IBC host instance if one is not already created
func (m *ibcModule) NewHost() (modules.IBCHost, error) {
	if m.host != nil {
		return nil, coreTypes.ErrHostAlreadyExists()
	}

	host := &host{
		logger: m.logger,
		stores: stores.NewStoreManager(m.storesDir),
	}

	m.host = host

	return host, nil
}
