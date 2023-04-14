package fisherman

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	FishermanModuleName = "fisherman"
)

type FishermanModule interface {
	modules.Module

	// TODO: add fisherman module functions here
	FishermanUtility
}

// TODO_IN_THIS_COMMIT: exists to help with type assertions, drop once you've added your own functions
type FishermanUtility interface {
	Fish()
}

type fisherman struct {
	base_modules.IntegratableModule
	logger *modules.Logger
}

// type assertions for fisherman module
var (
	_ FishermanModule = &fisherman{}
	// TODO: add fisherman module functions here
)

func CreateFisherman(bus modules.Bus, options ...modules.ModuleOption) (FishermanModule, error) {
	m, err := new(fisherman).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(FishermanModule), nil
}

func (*fisherman) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &fisherman{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

// TODO: implement this function
func (m *fisherman) Start() error {
	m.logger.Info().Msg("🎣 Fisherman module started 🎣")
	return nil
}

func (m *fisherman) Stop() error {
	m.logger.Info().Msg("🎣 Fisherman module stopped 🎣")
	return nil
}

func (m *fisherman) GetModuleName() string {
	return FishermanModuleName
}

func (m *fisherman) Fish() {
	m.logger.Info().Msg("🎣 Fisherman module is fishing 🎣")
}
