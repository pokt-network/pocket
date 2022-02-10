package modules

import (
	"log"

	"pocket/consensus/pkg/config"
	"pocket/shared/context"
)

// TODO: Show an example of `TypicalUsage`
type PocketModule interface {
	// func Create(ctx *context.PocketContext, base *modules.BasePocketModule) (interface{}, error) {}
	Start(*context.PocketContext) error
	Stop(*context.PocketContext) error

	SetPocketBusMod(PocketBusModule)
	GetPocketBusMod() PocketBusModule
}

type BasePocketModule struct {
	config       *config.Config
	pocketBusMod PocketBusModule

	// TODO: Create a custom logger for Pocket
	// TODO: Create a metrics module for Pocket
}

func NewBaseModule(
	ctx *context.PocketContext,
	config *config.Config,
) (*BasePocketModule, error) {
	return &BasePocketModule{
		config: config,
	}, nil
}

func (m *BasePocketModule) GetConfig() *config.Config {
	return m.config
}

func (m *BasePocketModule) SetPocketBusMod(pocketBus PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *BasePocketModule) GetPocketBusMod() PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}
