package persistance

import (
	"fmt"
	"log"
	"strconv"

	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
)

type persistanceModule struct {
	*modules.BasePocketModule
	modules.PersistanceModule
}

func Create(ctx *context.PocketContext, base *modules.BasePocketModule) (m modules.PersistanceModule, err error) {
	log.Println("Creating persistance module")
	m = &persistanceModule{
		BasePocketModule: base,
	}
	return m, nil
}

func (m *persistanceModule) Start(ctx *context.PocketContext) error {
	log.Println("Starting persistance module")
	return nil
}

func (m *persistanceModule) Stop(ctx *context.PocketContext) error {
	log.Println("Stopping persistance module")
	return nil
}

func (m *persistanceModule) GetLatestBlockHeight() (uint64, error) {
	log.Println("[TODO] Persistance GetLatestBlockHeight not implemented yet...")
	return 0, fmt.Errorf("PersistanceModule has not implemented GetLatestBlockHeight")
}

func (m *persistanceModule) GetBlockHash(height uint64) ([]byte, error) {
	log.Println("[TODO] Persistance GetLatestBlockHeight not implemented yet...")
	return []byte(strconv.FormatUint(height, 10)), nil
}

func (m *persistanceModule) GetPocketBusMod() modules.PocketBusModule {
	return m.BasePocketModule.GetPocketBusMod()
}

func (m *persistanceModule) SetPocketBusMod(bus modules.PocketBusModule) {
	m.BasePocketModule.SetPocketBusMod(bus)
}
