package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.PersistenceModule = &persistenceModule{}
	_ modules.PersistenceModule = &persistenceModule{}

	_ modules.PersistenceRWContext = &PostgresContext{}
)

// TODO: convert address and public key to string not bytes in all account and actor functions
// TODO: remove address parameter from all pool operations
type persistenceModule struct {
	bus          modules.Bus
	config       *configs.PersistenceConfig
	genesisState *genesis.GenesisState

	blockStore kvstore.KVStore
	txIndexer  indexer.TxIndexer
	stateTrees *stateTrees

	logger modules.Logger

	// TECHDEBT: Need to implement context pooling (for writes), timeouts (for read & writes), etc...
	writeContext *PostgresContext // only one write context is allowed at a time
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(persistenceModule).Create(bus)
}

func (*persistenceModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &persistenceModule{
		writeContext: nil,
	}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	persistenceCfg := runtimeMgr.GetConfig().Persistence
	genesisState := runtimeMgr.GetGenesis()

	conn, err := connectToDatabase(persistenceCfg)
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(conn); err != nil {
		return nil, err
	}
	conn.Close(context.TODO())

	// TODO: Follow the same pattern as txIndexer below for initializing the blockStore
	blockStore, err := initializeBlockStore(persistenceCfg.BlockStorePath)
	if err != nil {
		return nil, err
	}

	txIndexer, err := indexer.NewTxIndexer(persistenceCfg.TxIndexerPath)
	if err != nil {
		return nil, err
	}

	stateTrees, err := newStateTrees(persistenceCfg.TreesStoreDir)
	if err != nil {
		return nil, err
	}

	m.config = persistenceCfg
	m.genesisState = genesisState

	m.blockStore = blockStore
	m.txIndexer = txIndexer
	m.stateTrees = stateTrees

	// TECHDEBT: reconsider if this is the best place to call `populateGenesisState`. Note that
	// 		     this forces the genesis state to be reloaded on every node startup until state
	//           sync is implemented.
	// Determine if we should hydrate the genesis db or use the current state of the DB attached
	if shouldHydrateGenesis, err := m.shouldHydrateGenesisDb(); err != nil {
		return nil, err
	} else if shouldHydrateGenesis {
		m.populateGenesisState(genesisState) // fatal if there's an error
	} else {
		// This configurations will connect to the SQL database and key-value stores specified
		// in the configurations and connected to those.
		logger.Global.Info().Msg("Loading state from disk...")
	}

	return m, nil
}

func (m *persistenceModule) Start() error {
	m.logger.Info().Msg("Starting module...")
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())
	return nil
}

func (m *persistenceModule) Stop() error {
	m.blockStore.Stop()
	return nil
}

func (m *persistenceModule) GetModuleName() string {
	return modules.PersistenceModuleName
}

func (m *persistenceModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *persistenceModule) GetBus() modules.Bus {
	if m.bus == nil {
		logger.Global.Fatal().Msg("PocketBus is not initialized")
	}
	return m.bus
}

func (m *persistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	if m.writeContext != nil && !m.writeContext.conn.IsClosed() {
		return nil, fmt.Errorf("write context already exists")
	}
	conn, err := connectToDatabase(m.config)
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
		Height: height,
		conn:   conn,
		tx:     tx,

		stateHash: "",

		logger: m.logger,

		blockStore: m.blockStore,
		txIndexer:  m.txIndexer,
		stateTrees: m.stateTrees,
	}

	return m.writeContext, nil
}

func (m *persistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
	conn, err := connectToDatabase(m.config)
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
		conn:   conn,
		tx:     tx,

		stateHash: "",

		logger: m.logger,

		blockStore: m.blockStore,
		txIndexer:  m.txIndexer,
		stateTrees: m.stateTrees,
	}, nil
}

func (m *persistenceModule) ReleaseWriteContext() error {
	if m.writeContext != nil {
		if err := m.writeContext.resetContext(); err != nil {
			logger.Global.Error().Err(err).Msg("Error releasing write context")
		}
		m.writeContext = nil
	}
	return nil
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
//
//	move the if logic out of this file.
func (m *persistenceModule) shouldHydrateGenesisDb() (bool, error) {
	checkContext, err := m.NewReadContext(-1)
	if err != nil {
		return false, err
	}
	defer checkContext.Close()

	blockHeight, err := checkContext.GetLatestBlockHeight()
	if err != nil {
		return true, nil
	}

	if blockHeight == 0 {
		m.clearAllState(nil)
		return true, nil
	}

	return false, nil
}
