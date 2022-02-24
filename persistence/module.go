package persistence

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"

	"github.com/syndtr/goleveldb/leveldb/memdb"
)

type persistenceModule struct {
	modules.PersistenceModule
	pocketBus modules.Bus
}

func Create(cfg *config.Config) (modules.PersistenceModule, error) {
	return &persistenceModule{
		PersistenceModule: nil, // TODO(olshansky): sync with Andrew on a better way to do this
		pocketBus:         nil,
	}, nil

}

func (p *persistenceModule) Start() error {
	// TODO(olshansky): Add a test that pocketBus is set
	log.Println("Starting persistence module...")
	return nil
}

func (p *persistenceModule) Stop() error {
	log.Println("Stopping persistence module...")
	return nil
}

func (m *persistenceModule) SetBus(pocketBus modules.Bus) {
	m.pocketBus = pocketBus
}

func (m *persistenceModule) GetBus() modules.Bus {
	if m.pocketBus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBus
}

func (m *persistenceModule) NewContext(height int64) (modules.PersistenceContext, error) {
	panic("NewContext not implemented")
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	panic("GetCommitDB not implemented")
}
