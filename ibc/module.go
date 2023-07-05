package ibc

import (
	"context"

	"github.com/pokt-network/pocket/ibc/host"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.IBCModule = &ibcModule{}

type ibcModule struct {
	base_modules.IntegrableModule

	cfg    *configs.IBCConfig
	logger *modules.Logger

	// Only a single host is allowed at a time
	host modules.IBCHostSubmodule

	ctx    context.Context
	cancel context.CancelFunc
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(ibcModule).Create(bus, options...)
}

func (m *ibcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	*m = ibcModule{
		cfg:    bus.GetRuntimeMgr().GetConfig().IBC,
		logger: logger.Global.CreateLoggerForModule(modules.IBCModuleName),
	}
	m.logger.Info().Msg("🪐 creating IBC module 🪐")

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	// Only validators can be an IBC host due to the need for reliability
	isValidator := false
	if _, err := m.GetBus().GetUtilityModule().GetValidatorModule(); err == nil {
		isValidator = true
	}
	if isValidator && m.cfg.Enabled {
		if err := m.newHost(); err != nil {
			m.logger.Error().Err(err).Msg("❌ failed to create IBC host ❌")
			return nil, err
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancel = cancel

	return m, nil
}

func (m *ibcModule) Start() error {
	if !m.cfg.Enabled {
		m.logger.Info().Msg("🚫 IBC module disabled 🚫")
		return nil
	}
	if m.host != nil {
		if err := m.host.StartBackgroundTasks(m.ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *ibcModule) Stop() error {
	m.logger.Info().Msg("🛑 Stopping IBC Module 🛑")
	m.cancel()
	return nil
}

func (m *ibcModule) GetModuleName() string {
	return modules.IBCModuleName
}

// newHost creates a new IBC host and sets it in the ibcModule struct if it is not already set
func (m *ibcModule) newHost() error {
	if m.host != nil {
		return coreTypes.ErrIBCHostAlreadyExists()
	}
	hostMod, err := host.Create(m.GetBus(),
		m.cfg.Host,
		host.WithLogger(m.logger),
	)
	if err != nil {
		return err
	}
	m.host = hostMod
	return nil
}
