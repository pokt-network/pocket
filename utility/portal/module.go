package portal

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	PortalModuleName = "portal"
)

type PortalModule interface {
	modules.Module

	// TODO: add portal module functions here
	PortalUtility
}

// TODO_IN_THIS_COMMIT: exists to help with type assertions, drop once you've added your own functions
type PortalUtility interface {
	StartSession()
}

type portal struct {
	base_modules.IntegratableModule
	logger *modules.Logger
}

// type assertions for portal module
var (
	_ PortalModule = &portal{}
	// TODO: add portal module functions here
)

func CreatePortal(bus modules.Bus, options ...modules.ModuleOption) (PortalModule, error) {
	m, err := new(portal).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(PortalModule), nil
}

func (*portal) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &portal{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

// TODO: implement this function
func (m *portal) Start() error {
	m.logger.Info().Msg("🚪 Portal module started 🚪")
	return nil
}

func (m *portal) Stop() error {
	m.logger.Info().Msg("🚪 Portal module stopped 🚪")
	return nil
}

func (m *portal) GetModuleName() string {
	return PortalModuleName
}

func (m *portal) StartSession() {
	m.logger.Info().Msg("🚪 Portal module session started 🚪")
}
