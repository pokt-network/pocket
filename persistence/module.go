package persistence

import (
	"context"
	"fmt"
	"log"

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

type persistenceModule struct {
	bus          modules.Bus
	config       modules.PersistenceConfig
	genesisState modules.PersistenceGenesisState

	blockStore kvstore.KVStore // INVESTIGATE: We may need to create a custom `BlockStore` package in the future

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
	persistenceGenesis := genesis.GetPersistenceGenesisState().(*types.PersistenceGenesisState)

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
	// DISCUSS (team): we cannot cast if we want to use mocks and rely on interfaces
	// if _, ok := cfg.GetPersistenceConfig().(*types.PersistenceConfig); !ok {
	// 	 return fmt.Errorf("cannot cast to PersistenceConfig")
	// }
	return nil
}

func (*persistenceModule) ValidateGenesis(genesis modules.GenesisState) error {
	// DISCUSS (team): we cannot cast if we want to use mocks and rely on interfaces
	// if _, ok := genesis.GetPersistenceGenesisState().(*types.PersistenceGenesisState); !ok {
	// 	return fmt.Errorf("cannot cast to PersistenceGenesisState")
	// }
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
	}

	return *m.writeContext, nil

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
	}, nil
}

func (m *persistenceModule) ResetContext() error {
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

func initializeBlockStore(blockStorePath string) (kvstore.KVStore, error) {
	if blockStorePath == "" {
		return kvstore.NewMemKVStore(), nil
	}
	return kvstore.NewKVStore(blockStorePath)
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

	return m.blockStore.Exists(heightToBytes(int64(maxHeight)))
}
