package persistence

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

var _ modules.PersistenceModule = &persistenceModule{}
var _ modules.PersistenceRWContext = &PostgresContext{}

type persistenceModule struct {
	bus modules.Bus

	postgresURL string
	nodeSchema  string
	blockStore  kvstore.KVStore // INVESTIGATE: We may need to create a custom `BlockStore` package in the future

	// TECHDEBT: Need to implement context pooling (for writes), timeouts (for read & writes), etc...
	writeContext *PostgresContext // only one write context is allowed at a time
}

func Create(cfg *genesis.Config, genesis *genesis.GenesisState) (modules.PersistenceModule, error) {
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
		bus:          nil,
		postgresURL:  cfg.Persistence.PostgresUrl,
		nodeSchema:   cfg.Persistence.NodeSchema,
		blockStore:   blockStore,
		writeContext: nil,
	}

	// Determine if we should hydrate the genesis db or use the current state of the DB attached
	if shouldHydrateGenesis, err := persistenceMod.shouldHydrateGenesisDb(); err != nil {
		return nil, err
	} else if shouldHydrateGenesis {
		// TECHDEBT: reconsider if this is the best place to call `populateGenesisState`. Note that
		// 		     this forces the genesis state to be reloaded on every node startup until state sync is
		//           implemented.
		// NOTE: `populateGenesisState` does not return an error but logs a fatal error if there's a problem
		persistenceMod.populateGenesisState(genesis)
	} else {
		log.Println("Loading state from previous state...")
	}

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

func (m *persistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	if m.writeContext != nil && !m.writeContext.DB.conn.IsClosed() {
		return nil, fmt.Errorf("write context already exists")
	}
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

	m.writeContext = &PostgresContext{
		Height: height,
		DB: PostgresDB{
			conn:       conn,
			Tx:         tx,
			Blockstore: m.blockStore,
		},
	}

	return *m.writeContext, nil

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

func (m *persistenceModule) ResetContext() error {
	if m.writeContext != nil {
		if !m.writeContext.DB.Tx.Conn().IsClosed() {
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
