package utility

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"
)

var _ modules.UtilityModule = &utilityModule{}

type utilityModule struct {
	modules.UtilityModule
	pocketBus modules.Bus
}

func Create(cfg *config.Config) (modules.UtilityModule, error) {
	m := &utilityModule{
		UtilityModule: nil, // TODO(olshansky): sync with Andrew on a better way to do this
		pocketBus:     nil,
	}
	return m, nil
}

func (u *utilityModule) Start() error {
	// TODO(olshansky): Add a test that pocketBus is set
	log.Println("Starting consensus module...")
	return nil
}

func (u *utilityModule) Stop() error {
	log.Println("Stopping consensus module...")
	return nil
}

func (u *utilityModule) GetBus() modules.Bus {
	if u.pocketBus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return u.pocketBus
}

func (u *utilityModule) SetBus(pocketBus modules.Bus) {
	u.pocketBus = pocketBus
}

func (u *utilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	panic("NewContext not implemented")
}
