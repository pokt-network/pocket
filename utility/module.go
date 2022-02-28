package utility

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"
)

var _ modules.UtilityModule = &utilityModule{}

type utilityModule struct {
	bus modules.Bus
}

func Create(cfg *config.Config) (modules.UtilityModule, error) {
	m := &utilityModule{
		bus: nil,
	}
	return m, nil
}

func (u *utilityModule) Start() error {
	// TODO(olshansky): Add a test that bus is set
	log.Println("Starting utility module...")
	return nil
}

func (u *utilityModule) Stop() error {
	log.Println("Stopping utility module...")
	return nil
}

func (u *utilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return u.bus
}

func (u *utilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *utilityModule) NewContext(height int64) (modules.UtilityContext, error) {
	panic("NewContext not implemented")
}
