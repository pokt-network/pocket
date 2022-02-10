package persistence

import (
	"fmt"
	"log"
	"pocket/consensus/pkg/config"
	"strconv"

	"pocket/shared/context"
	"pocket/shared/modules"
)

type persistenceModule struct {
	modules.PersistenceModule
	pocketBusMod modules.PocketBusModule
}

func Create(ctx *context.PocketContext) (m modules.PersistenceModule, err error) {
	log.Println("Creating persistence module")
	m = &persistenceModule{}
	return m, nil
}

func (m *persistenceModule) Start(ctx *context.PocketContext, cfg *config.Config) error {
	log.Println("Starting persistence module")
	return nil
}

func (m *persistenceModule) Stop(ctx *context.PocketContext) error {
	log.Println("Stopping persistence module")
	return nil
}

func (m *persistenceModule) GetLatestBlockHeight() (uint64, error) {
	log.Println("[TODO] persistence GetLatestBlockHeight not implemented yet...")
	return 0, fmt.Errorf("persistenceModule has not implemented GetLatestBlockHeight")
}

func (m *persistenceModule) GetBlockHash(height uint64) ([]byte, error) {
	log.Println("[TODO] persistence GetLatestBlockHeight not implemented yet...")
	return []byte(strconv.FormatUint(height, 10)), nil
}