package persistence

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.PersistenceModule = &persistenceModule{}
var _ modules.PersistenceRWContext = &PostgresContext{}

type persistenceModule struct {
	bus modules.Bus

	db          *pgx.Conn
	postgresURL string
	nodeSchema  string
	blockStore  kvstore.KVStore // INVESTIGATE: We may need to create a custom `BlockStore` package in the future
}

func Create(c *config.Config) (modules.PersistenceModule, error) {
	db, err := connectAndInitializeDatabase(c.Persistence.PostgresUrl, c.Persistence.NodeSchema)
	if err != nil {
		return nil, err
	}
	blockStore, err := initializeBlockStore(c.Persistence.BlockStorePath)
	if err != nil {
		return nil, err
	}
	return &persistenceModule{
		postgresURL: c.Persistence.PostgresUrl,
		nodeSchema:  c.Persistence.NodeSchema,
		db:          db,
		bus:         nil,
		blockStore:  blockStore,
	}, nil
}

func (p *persistenceModule) Start() error {
	log.Println("Starting persistence module...")
	// DISCUSS_IN_THIS_COMMIT: When should this run?
	p.populateGenesisState(p.bus.GetConfig().GenesisSource.GetState())
	return nil
}

func (p *persistenceModule) Stop() error {
	p.blockStore.Stop()
	p.db.Close(context.TODO())
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

func (m *persistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	db, err := connectAndInitializeDatabase(m.postgresURL, m.nodeSchema)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return nil, err
	}

	return PostgresContext{
		Height: height,
		DB: PostgresDB{
			Tx:         tx,
			Blockstore: m.blockStore,
		},
	}, nil
}

func (m *persistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
	return m.NewRWContext(height)
	// TODO (Team) this can be completely separate from rw context.
	// It should access the db directly rather than using transactions
}

func (m *persistenceModule) GetBlockStore() kvstore.KVStore {
	return m.blockStore
}

func initializeBlockStore(blockStorePath string) (kvstore.KVStore, error) {
	if blockStorePath == "" {
		return kvstore.NewMemKVStore(), nil
	}
	return kvstore.NewKVStore(blockStorePath)
}
