package persistence

import (
	"context"
	"fmt"
	"log"

	"github.com/pokt-network/pocket/persistence/indexer"

	"github.com/pokt-network/pocket/persistence/types"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.PersistenceModule = &persistenceModule{}
	_ modules.PersistenceModule = &persistenceModule{}

	_ modules.PersistenceRWContext    = &PostgresContext{}
	_ modules.PersistenceGenesisState = &types.PersistenceGenesisState{}
	_ modules.PersistenceConfig       = &types.PersistenceConfig{}
)

// TODO: convert address and public key to string not bytes in all account and actor functions
// TODO: remove address parameter from all pool operations
type persistenceModule struct {
	bus          modules.Bus
	config       modules.PersistenceConfig
	genesisState modules.PersistenceGenesisState

	blockStore kvstore.KVStore // INVESTIGATE: We may need to create a custom `BlockStore` package in the future
	txIndexer  indexer.TxIndexer

	// TECHDEBT: Need to implement context pooling (for writes), timeouts (for read & writes), etc...
	writeContext *PostgresContext // only one write context is allowed at a time
}

const (
	PersistenceModuleName = "persistence"
)

func Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return new(persistenceModule).Create(runtimeMgr)
}

func (*persistenceModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m *persistenceModule

	cfg := runtimeMgr.GetConfig()

	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	persistenceCfg := cfg.GetPersistenceConfig()

	genesis := runtimeMgr.GetGenesis()
	if err := m.ValidateGenesis(genesis); err != nil {
		return nil, fmt.Errorf("genesis validation failed: %w", err)
	}
	persistenceGenesis := genesis.GetPersistenceGenesisState()

	conn, err := connectToDatabase(persistenceCfg.GetPostgresUrl(), persistenceCfg.GetNodeSchema())
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(conn); err != nil {
		return nil, err
	}
	conn.Close(context.TODO())

	blockStore, err := initializeBlockStore(persistenceCfg.GetBlockStorePath())
	if err != nil {
		return nil, err
	}

	txIndexer, err := indexer.NewTxIndexer(persistenceCfg.GetTxIndexerPath())
	if err != nil {
		return nil, err
	}

	m = &persistenceModule{
		bus:          nil,
		config:       persistenceCfg,
		genesisState: persistenceGenesis,
		blockStore:   blockStore,
		txIndexer:    txIndexer,
		writeContext: nil,
	}

	// Determine if we should hydrate the genesis db or use the current state of the DB attached
	if shouldHydrateGenesis, err := m.shouldHydrateGenesisDb(); err != nil {
		return nil, err
	} else if shouldHydrateGenesis {
		// TECHDEBT: reconsider if this is the best place to call `populateGenesisState`. Note that
		// 		     this forces the genesis state to be reloaded on every node startup until state sync is
		//           implemented.
		// NOTE: `populateGenesisState` does not return an error but logs a fatal error if there's a problem
		m.populateGenesisState(persistenceGenesis)
	} else {
		log.Println("Loading state from previous state...")
	}

	return m, nil
}

func (m *persistenceModule) Start() error {
	log.Println("Starting persistence module...")
	return nil
}

func (m *persistenceModule) Stop() error {
	m.blockStore.Stop()
	return nil
}

func (m *persistenceModule) GetModuleName() string {
	return PersistenceModuleName
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

func (*persistenceModule) ValidateConfig(cfg modules.Config) error {
	return nil
}

func (*persistenceModule) ValidateGenesis(genesis modules.GenesisState) error {
	return nil
}

func (m *persistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	if m.writeContext != nil && !m.writeContext.conn.IsClosed() {
		return nil, fmt.Errorf("write context already exists")
	}
	conn, err := connectToDatabase(m.config.GetPostgresUrl(), m.config.GetNodeSchema())
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

	m.writeContext = &PostgresContext{
		Height:     height,
		conn:       conn,
		tx:         tx,
		blockstore: m.blockStore,
		txIndexer:  m.txIndexer,
	}

	return m.writeContext, nil

}

func (m *persistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
	conn, err := connectToDatabase(m.config.GetPostgresUrl(), m.config.GetNodeSchema())
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
		Height:     height,
		conn:       conn,
		tx:         tx,
		blockstore: m.blockStore,
		txIndexer:  m.txIndexer,
	}, nil
}

func (m *persistenceModule) GetBlockStore() kvstore.KVStore {
	return m.blockStore
}

func (m *persistenceModule) NewWriteContext() modules.PersistenceRWContext {
	return m.writeContext
}

func initializeBlockStore(blockStorePath string) (kvstore.KVStore, error) {
	if blockStorePath == "" {
		return kvstore.NewMemKVStore(), nil
	}
	return kvstore.NewKVStore(blockStorePath)
}

// HACK(olshansky): Simplify and externalize the logic for whether genesis should be populated and
//                  move the if logic out of this file.
func (m *persistenceModule) shouldHydrateGenesisDb() (bool, error) {
	checkContext, err := m.NewReadContext(-1)
	if err != nil {
		return false, err
	}
	defer checkContext.Close()

	if _, err = checkContext.GetLatestBlockHeight(); err != nil {
	if err != nil {
		return true, nil
	}
	return false, nil
}
