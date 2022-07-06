package persistence

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"

	"github.com/syndtr/goleveldb/leveldb/memdb"
)

var _ modules.PersistenceModule = &persistenceModule{}

type persistenceModule struct {
	bus modules.Bus
}

func Create(c *config.Config) (modules.PersistenceModule, error) {
	if _, err := ConnectAndInitializeDatabase(c.Persistence.PostgresUrl, c.Persistence.NodeSchema); err != nil {
		return nil, err
	}
	return &persistenceModule{
		bus: nil,
	}, nil
}

func (p *persistenceModule) Start() error {
	log.Println("Starting persistence module...")
	return nil
}

func (p *persistenceModule) Stop() error {
	log.Println("Stopping persistence module...")
	return nil
}

func (m *persistenceModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *persistenceModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *persistenceModule) NewContext(height int64) (modules.PersistenceContext, error) {
	panic("NewContext not implemented")
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	panic("GetCommitDB not implemented")
}
