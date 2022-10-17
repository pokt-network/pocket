package persistence

import (
	"context"
	"fmt"
	"log"

	"github.com/celestiaorg/smt"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
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

	// postgresURL string
	// nodeSchema  string
	// genesisPath string

	// TECHDEBT: Need to implement context pooling (for writes), timeouts (for read & writes), etc...
	writeContext *PostgresContext // only one write context is allowed at a time

	// The connection to the PostgreSQL database
	postgresConn *pgx.Conn
	// A reference to the block key-value store
	// INVESTIGATE: We may need to create a custom `BlockStore` package in the future.
	blockStore kvstore.KVStore
	// A mapping of context IDs to persistence contexts
	// contexts map[contextId]modules.PersistenceRWContext
	// Merkle trees
	trees map[MerkleTree]*smt.SparseMerkleTree
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

	m = &persistenceModule{
		bus:          nil,
		config:       persistenceCfg,
		genesisState: persistenceGenesis,
		blockStore:   blockStore,
		writeContext: nil,
		// contexts:     make(map[contextId]modules.PersistenceContext),
		trees: make(map[MerkleTree]*smt.SparseMerkleTree),
	}

	// DISCUSS_IN_THIS_COMMIT: We've been using the module function pattern, but this making `initializeTrees`
	// be able to create and/or load trees outside the scope of the persistence module makes it easier to test.
	trees, err := newMerkleTrees()
	if err != nil {
		return nil, err
	}

	// TODO_IN_THIS_COMMIT: load trees from state
	m.trees = trees

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
		Height: height,
		conn:   conn,
		tx:     tx,

		currentBlockTxs:  make([][]byte, 0),
		currentStateHash: make([]byte, 0),

		blockStore:  m.blockStore,
		merkleTrees: m.trees,
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
		blockStore: m.blockStore,
	}, nil
}

func (m *persistenceModule) ReleaseWriteContext() error {
	if m.writeContext != nil {
		if !m.writeContext.GetTx().Conn().IsClosed() {
			if err := m.writeContext.Release(); err != nil {
				log.Println("[TODO][ERROR] Error releasing write context...", err)
			}
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
	return kvstore.OpenKVStore(blockStorePath)
}

// TODO(drewsky): Simplify and externalize the logic for whether genesis should be populated and
// move the if logic out of this file.
func (m *persistenceModule) shouldHydrateGenesisDb() (bool, error) {
	checkContext, err := m.NewReadContext(-1)
	if err != nil {
		return false, err
	}
	defer checkContext.Close()

	maxHeight, err := checkContext.GetLatestBlockHeight()
	if err == nil || maxHeight == 0 {
		return true, nil
	}

	return m.blockStore.Exists(HeightToBytes(int64(maxHeight)))
}
