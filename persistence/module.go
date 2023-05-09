package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence/blockstore"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ modules.PersistenceModule = &persistenceModule{}
	_ modules.PersistenceModule = &persistenceModule{}

	_ modules.PersistenceRWContext = &PostgresContext{}
)

// TODO: convert address and public key to string not bytes in all account and actor functions
// TODO: remove address parameter from all pool operations
type persistenceModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger

	config       *configs.PersistenceConfig
	genesisState *genesis.GenesisState
	networkId    string

	// A key-value store mapping heights to blocks. Needed for block synchronization.
	blockStore *blockstore.BlockStore

	// A tx indexer (i.e. key-value store) mapping transaction hashes to transactions. Needed for
	// avoiding tx replays attacks, and is also used as the backing database for the transaction
	// tx merkle tree.
	txIndexer indexer.TxIndexer

	// A list of all the merkle trees maintained by the persistence module that roll up into the state commitment.
	stateTrees *stateTrees

	// Only one write context is allowed at a time
	writeContext *PostgresContext

	// A pool of connections to the postgres database
	pool *pgxpool.Pool
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(persistenceModule).Create(bus, options...)
}

func (*persistenceModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &persistenceModule{
		writeContext: nil,
	}

	// TECHDEBT: move to `persistenceModule#Start` as per documentation.
	// Temporarily moving this here as long as there are references to
	// the logger in methods which are called by `#Create` (i.e.
	// `persistenceModule#populateGenesisState`, `postgresContext#Commit`)
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	persistenceCfg := runtimeMgr.GetConfig().Persistence
	genesisState := runtimeMgr.GetGenesis()

	pool, err := initializePool(persistenceCfg)
	if err != nil {
		return nil, err
	}
	m.pool = pool

	conn, err := connectToPool(m.pool, persistenceCfg.GetNodeSchema())
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(conn); err != nil {
		return nil, err
	}
	conn.Release()

	blockStore, err := blockstore.NewBlockStore(persistenceCfg.BlockStorePath)
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
	m.networkId = runtimeMgr.GetConfig().NetworkId

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
	return nil
}

func (m *persistenceModule) Stop() error {
	m.pool.Close()
	return m.blockStore.Stop()
}

func (m *persistenceModule) GetModuleName() string {
	return modules.PersistenceModuleName
}

func (m *persistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	if m.writeContext != nil && m.writeContext.isOpen() {
		return nil, fmt.Errorf("cannot create a new write context if one already exists")
	}

	// Take one of the connections from the db pool
	conn, err := connectToPool(m.pool, m.config.GetNodeSchema())
	if err != nil {
		return nil, err
	}

	// Start a new database transaction
	tx, err := conn.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadUncommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.Deferrable, // INVESTIGATE: Research if this should be `Deferrable`
	})
	if err != nil {
		return nil, err
	}

	m.writeContext = &PostgresContext{
		logger: m.logger,
		Height: height,

		conn: conn,
		tx:   tx,

		stateHash:  "",
		blockStore: m.blockStore,
		txIndexer:  m.txIndexer,
		stateTrees: m.stateTrees,
		networkId:  m.networkId,
	}

	return m.writeContext, nil
}

func (m *persistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
	conn, err := connectToPool(m.pool, m.config.GetNodeSchema())
	if err != nil {
		return nil, err
	}
	tx, err := conn.BeginTx(context.TODO(), pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadOnly,
		DeferrableMode: pgx.NotDeferrable, // INVESTIGATE: Research if this should be `Deferrable`
	})
	if err != nil {
		return nil, err
	}

	return &PostgresContext{
		logger: m.logger,

		Height: height,

		conn: conn,
		tx:   tx,

		stateHash:  "",
		blockStore: m.blockStore,
		txIndexer:  m.txIndexer,
		stateTrees: m.stateTrees,
		networkId:  m.networkId,
	}, nil
}

func (m *persistenceModule) ReleaseWriteContext() error {
	writeContext := m.writeContext
	if writeContext == nil {
		return nil
	}
	m.writeContext = nil
	if !writeContext.isOpen() {
		return nil
	}
	writeContext.Release()
	return nil
}

func (m *persistenceModule) GetBlockStore() *blockstore.BlockStore {
	return m.blockStore
}

func (m *persistenceModule) GetTxIndexer() indexer.TxIndexer {
	return m.txIndexer
}

func (m *persistenceModule) GetNetworkID() string {
	return m.networkId
}

func (m *persistenceModule) NewWriteContext() modules.PersistenceRWContext {
	return m.writeContext
}

// HACK(olshansky): Simplify and externalize the logic for whether genesis should be populated and
//
//	move the if logic out of this file.
func (m *persistenceModule) shouldHydrateGenesisDb() (bool, error) {
	readCtx, err := m.NewReadContext(-1)
	if err != nil {
		return false, err
	}
	defer readCtx.Release()

	blockHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil {
		return true, nil
	}

	if blockHeight > 0 {
		return false, nil
	}

	if err := m.clearAllState(nil); err != nil {
		return false, err
	}

	return true, nil
}
