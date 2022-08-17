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

	postgresURL string
	nodeSchema  string
	blockStore  kvstore.KVStore // INVESTIGATE: We may need to create a custom `BlockStore` package in the future
}

func Create(cfg *config.Config) (modules.PersistenceModule, error) {
	conn, err := connectToDatabase(cfg.Persistence.PostgresUrl, cfg.Persistence.NodeSchema)
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(conn); err != nil {
		return nil, err
	}
	conn.Close(context.TODO())

	blockStore, err := initializeBlockStore(cfg.Persistence.BlockStorePath)
	if err != nil {
		return nil, err
	}

	persistenceMod := &persistenceModule{
		bus:         nil,
		postgresURL: cfg.Persistence.PostgresUrl,
		nodeSchema:  cfg.Persistence.NodeSchema,
		blockStore:  blockStore,
	}

	// TECHDEBT: reconsider if this is the best place to call `populateGenesisState`. Note that
	// this forces the genesis state to be reloaded on every node startup until state sync is
	// implemented.
	persistenceMod.populateGenesisState(cfg.GenesisSource.GetState())

	return persistenceMod, nil
}

func (m *persistenceModule) Start() error {
	log.Println("Starting persistence module...")
	return nil
}

func (m *persistenceModule) Stop() error {
	m.blockStore.Stop()
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
	conn, err := connectToDatabase(m.postgresURL, m.nodeSchema)
	if err != nil {
		return nil, err
	}
	tx, err := conn.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadUncommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.Deferrable, // TODO(andrew): Research if this should be `Deferrable`
	})
	if err != nil {
		return nil, err
	}
	return PostgresContext{
		Height: height,
		DB: PostgresDB{
			conn:       conn,
			Tx:         tx,
			Blockstore: m.blockStore,
		},
	}, nil
}

func (m *persistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
	conn, err := connectToDatabase(m.postgresURL, m.nodeSchema)
	if err != nil {
		return nil, err
	}

	tx, err := conn.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadOnly,
		DeferrableMode: pgx.NotDeferrable, // TODO(andrew): Research if this should be `Deferrable`
	})
	if err != nil {
		return nil, err
	}

	return PostgresContext{
		Height: height,
		DB: PostgresDB{
			conn:       conn,
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
