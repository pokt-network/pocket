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
var _ modules.PersistenceContext = &PostgresContext{}

func (p PostgresContext) GetAppStakeAmount(height int64, address []byte) (string, error) {
	panic("TODO: implement PostgresContext.GetAppStakeAmount")
}

func (p PostgresContext) SetAppStakeAmount(address []byte, stakeAmount string) error {
	panic("TODO: implement PostgresContext.SetAppStakeAmount")
}

func (p PostgresContext) GetServiceNodeStakeAmount(height int64, address []byte) (string, error) {
	panic("TODO: implement PostgresContext.GetServiceNodeStakeAmount")
}

func (p PostgresContext) SetServiceNodeStakeAmount(address []byte, stakeAmount string) error {
	panic("TODO: implement PostgresContext.SetServiceNodeStakeAmount")
}

func (p PostgresContext) GetFishermanStakeAmount(height int64, address []byte) (string, error) {
	panic("TODO: implement PostgresContext.SetServiceNodeStakeAmount")
}

func (p PostgresContext) SetFishermanStakeAmount(address []byte, stakeAmount string) error {
	panic("TODO: implement PostgresContext.SetFishermanStakeAmount")
}

func (p PostgresContext) GetValidatorStakeAmount(height int64, address []byte) (string, error) {
	panic("TODO: implement PostgresContext.GetValidatorStakeAmount")
}

func (p PostgresContext) SetValidatorStakeAmount(address []byte, stakeAmount string) error {
	panic("TODO: implement PostgresContext.SetValidatorStakeAmount")
}

type persistenceModule struct {
	bus modules.Bus

	// The connection to the PostgreSQL database
	postgresConn *pgx.Conn
	// A reference to the block key-value store
	// INVESTIGATE: We may need to create a custom `BlockStore` package in the future.
	blockStore kvstore.KVStore
	// A mapping of context IDs to persistence contexts
	contexts map[contextId]modules.PersistenceContext
}

type contextId uint64

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
		contexts:     make(map[contextId]modules.PersistenceContext),
	}, nil
}

func (p *persistenceModule) Start() error {
	log.Println("Starting persistence module...")

	shouldHydrateGenesis := false
	shouldHydrateGenesis, err := p.shouldHydrateGenesisDb()
	if err != nil {
		return err
	}

	if shouldHydrateGenesis {
		if err := p.hydrateGenesisDbState(); err != nil {
			return err
		}
		log.Println("Hydrating genesis state...")
	} else {
		log.Println("Loading state from previous state...")
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
	persistenceContext := PostgresContext{
		Height:       height,
		PostgresDB:   m.postgresConn,
		BlockStore:   m.blockStore,
		ContextStore: kvstore.NewMemKVStore(),
	}

	m.contexts[createContextId(height)] = persistenceContext

	return persistenceContext, nil
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	panic("GetCommitDB not implemented")
}

func (m *persistenceModule) GetBlockStore() kvstore.KVStore {
	return m.blockStore
}

// INCOMPLETE: We will need to support multiple contexts at the same height in the future
func createContextId(height int64) contextId {
	return contextId(height)
}

// INCOMPLETE: This is not a complete implementation but just a first approach. Approach with
//             a grain of salt.
func (m *persistenceModule) shouldHydrateGenesisDb() (bool, error) {
	checkContext, err := m.NewContext(-1) // Unknown height
	if err != nil {
		return false, err
	}
	defer checkContext.Release()

	maxHeight, err := checkContext.GetLatestBlockHeight()
	if err == nil || maxHeight == 0 {
		return true, nil
	}

	return m.blockStore.Exists(heightToBytes(int64(maxHeight)))
}
