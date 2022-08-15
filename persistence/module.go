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

func Create(cfg *config.Config) (modules.PersistenceModule, error) {
	db, err := connectAndInitializeDatabase(cfg.Persistence.PostgresUrl, cfg.Persistence.NodeSchema)
	if err != nil {
		return nil, err
	}
	blockStore, err := initializeBlockStore(cfg.Persistence.BlockStorePath)
	if err != nil {
		return nil, err
	}
	persistenceMod := &persistenceModule{
		postgresURL: cfg.Persistence.PostgresUrl,
		nodeSchema:  cfg.Persistence.NodeSchema,
		db:          db,
		bus:         nil,
		blockStore:  blockStore,
	}
	// DISCUSS_IN_THIS_COMMIT: Is `Create` the appropriate location for this or should it be `Start`?
	// DISCUSS_IN_THIS_COMMIT: Thoughts on bringing back `shouldHydrateGenesisDb`? It allowed LocalNet
	//                         to continue from a previously stored state.
	persistenceMod.populateGenesisState(cfg.GenesisSource.GetState())

	return persistenceMod, nil
}

func (m *persistenceModule) Start() error {
	log.Println("Starting persistence module...")
	return nil
}

func (m *persistenceModule) Stop() error {
	m.blockStore.Stop()
	m.db.Close(context.TODO())
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

// TECHDEBT: Only one write context at a time should be allowed
func (m *persistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	tx, err := m.db.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadUncommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable, // DISCUSS_IN_THIS_COMMIT: Should this be Deferrable?
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
	tx, err := m.db.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadOnly,
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

func (m *persistenceModule) GetBlockStore() kvstore.KVStore {
	return m.blockStore
}

func initializeBlockStore(blockStorePath string) (kvstore.KVStore, error) {
	if blockStorePath == "" {
		return kvstore.NewMemKVStore(), nil
	}
	return kvstore.NewKVStore(blockStorePath)
}
