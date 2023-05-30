package ibc

import (
	"github.com/pokt-network/pocket/ibc/stores"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.IBCModule = (*ibcModule)(nil)

type ibcModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	// Only a single host is allowed at a time
	host *Host
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

	return m, nil
}

func (m *ibcModule) Start() error {
	return nil
}

func (m *ibcModule) Stop() error {
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

	host := &Host{
		logger: m.logger,
		stores: stores.NewStoreManager(),
	}

	m.host = host

	return host, nil
}
