package fisherman

import (
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

const (
	FishermanModuleName = "fisherman"
)

type fisherman struct {
	base_modules.IntegrableModule
	logger *modules.Logger
}

var (
	_ modules.FishermanModule = &fisherman{}
)

func CreateFisherman(bus modules.Bus, options ...modules.ModuleOption) (modules.FishermanModule, error) {
	m, err := new(fisherman).Create(bus, options...)
	if err != nil {
		return nil, err
	}
	return m.(modules.FishermanModule), nil
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
