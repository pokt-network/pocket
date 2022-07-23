package persistence

import (
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"

	"github.com/syndtr/goleveldb/leveldb/memdb"
)

var _ modules.PersistenceModule = &persistenceModule{}

type persistenceModule struct {
	bus modules.Bus

	postgresConn *pgx.Conn
	blockStore   kvstore.KVStore
	contexts     map[uint64]modules.PersistenceContext
}

func Create(c *config.Config) (modules.PersistenceModule, error) {
	postgresDb, err := ConnectAndInitializeDatabase(c.Persistence.PostgresUrl, c.Persistence.NodeSchema)
	if err != nil {
		return nil, err
	}

	blockStore, err := kvstore.NewKVStore(c.Persistence.BlockStorePath)
	if err != nil {
		return nil, err
	}

	return &persistenceModule{
		bus: nil,

		postgresConn: postgresDb,
		blockStore:   blockStore,
		contexts:     make(map[uint64]modules.PersistenceContext),
	}, nil
}

func (p *persistenceModule) Start() error {
	log.Println("Starting persistence module...")

	// TODO: Load from previous state
	if err := p.hydrateGenesisDbState(); err != nil {
		return err
	}

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
	// persistenceContext, err := CreatePersistenceContext()
	// if err != nil {
	// 	return nil, err
	// }

	persistenceContext := PostgresContext{
		Height: height,
		DB: &PostgresDB{
			Conn: m.postgresConn,
		},
	}

	m.contexts[uint64(height)] = persistenceContext

	// return CreatePersistenceContext()

	// newDB := NewMemDB()
	// it := m.CommitDB.NewIterator(&util.Range{
	// 	Start: HeightKey(height, nil),
	// 	Limit: HeightKey(height+1, nil),
	// })
	// defer it.Release()
	// for valid := it.First(); valid; valid = it.Next() {
	// 	err := newDB.Put(KeyFromHeightKey(it.Key()), it.Value())
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	// context := &PrePersistenceContext{
	// 	Height: height,
	// 	Parent: m,
	// 	DBs:    make([]*memdb.DB, 0),
	// }
	// context.DBs = append(context.DBs, newDB)
	return persistenceContext, nil
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	panic("GetCommitDB not implemented")
}
