package servicer

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	ServicerModuleName = "servicer"
)

type ServicerModule interface {
	modules.Module
	ServicerUtility
}

type ServicerUtility interface{}

type servicer struct {
	base_modules.IntegratableModule
	logger *modules.Logger
}

// type assertions for servicer module
var (
	_ ServicerModule = &servicer{}
)

func CreateServicer(bus modules.Bus, options ...modules.ModuleOption) (ServicerModule, error) {
	m, err := new(servicer).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(ServicerModule), nil
}

func (*servicer) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &servicer{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

// TODO: implement this function
func (m *servicer) Start() error {
	m.logger.Info().Msg("🧬 Servicer module started 🧬")
	return nil
}

func (m *servicer) Stop() error {
	m.logger.Info().Msg("🧬 Servicer module stopped 🧬")
	return nil
}

func (m *servicer) GetModuleName() string {
	return ServicerModuleName
}
