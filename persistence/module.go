package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/celestiaorg/smt"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/test_artifacts"
)

var _ modules.PersistenceModule = &PersistenceModule{}
var _ modules.PersistenceRWContext = &PostgresContext{}
var _ modules.PersistenceGenesisState = &types.PersistenceGenesisState{}
var _ modules.PersistenceConfig = &types.PersistenceConfig{}

type PersistenceModule struct {
	bus modules.Bus

	postgresURL string
	nodeSchema  string
	genesisPath string

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

func Create(configPath, genesisPath string) (modules.PersistenceModule, error) {
	m := new(PersistenceModule)
	c, err := m.InitConfig(configPath)
	if err != nil {
		return nil, err
	}

	cfg := c.(*types.PersistenceConfig)
	g, err := m.InitGenesis(genesisPath)

	if err != nil {
		return nil, err
	}
	genesis := g.(*types.PersistenceGenesisState)
	conn, err := connectToDatabase(cfg.GetPostgresUrl(), cfg.GetNodeSchema())
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(conn); err != nil {
		return nil, err
	}
	conn.Close(context.TODO())

	blockStore, err := initializeBlockStore(cfg.GetBlockStorePath())
	if err != nil {
		return nil, err
	}

	persistenceMod := &PersistenceModule{
		bus:          nil,
		postgresURL:  cfg.GetPostgresUrl(),
		nodeSchema:   cfg.GetNodeSchema(),
		genesisPath:  genesisPath,
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
	persistenceMod.trees = trees

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

func (m *PersistenceModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the %s file: %v", pathToConfigJSON, err.Error())
	}
	// persistence specific configuration file
	config = new(types.PersistenceConfig)
	err = json.Unmarshal(rawJSON[m.GetModuleName()], config)
	return
}

func (m *PersistenceModule) InitGenesis(pathToGenesisJSON string) (genesis modules.IGenesis, err error) {
	data, err := ioutil.ReadFile(pathToGenesisJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the %s file: %v", pathToGenesisJSON, err.Error())
	}
	// persistence specific configuration file
	genesis = new(types.PersistenceGenesisState)
	err = json.Unmarshal(rawJSON[test_artifacts.GetGenesisFileName(m.GetModuleName())], genesis)
	return
}

func (m *PersistenceModule) Start() error {
	log.Println("Starting persistence module...")
	return nil
}

func (m *PersistenceModule) Stop() error {
	m.blockStore.Stop()
	return nil
}

func (m *PersistenceModule) GetModuleName() string {
	return PersistenceModuleName
}

func (m *PersistenceModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *PersistenceModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *PersistenceModule) NewRWContext(height int64) (modules.PersistenceRWContext, error) {
	if m.writeContext != nil && !m.writeContext.conn.IsClosed() {
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
		conn:   conn,
		tx:     tx,

		currentBlockTxs:  make([][]byte, 0),
		currentStateHash: make([]byte, 0),

		blockStore:  m.blockStore,
		merkleTrees: m.trees,
	}

	return m.writeContext, nil

}

func (m *PersistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
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
		Height:     height,
		conn:       conn,
		tx:         tx,
		blockStore: m.blockStore,
	}, nil
}

func (m *PersistenceModule) ReleaseWriteContext() error {
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

func (m *PersistenceModule) GetBlockStore() kvstore.KVStore {
	return m.blockStore
}

func initializeBlockStore(blockStorePath string) (kvstore.KVStore, error) {
	if blockStorePath == "" {
		return kvstore.NewMemKVStore(), nil
	}
	return kvstore.OpenKVStore(blockStorePath)
}

// TODO(drewsky): Simplify and externalize the logic for whether genesis should be populated and
// move the if logic out of this file.
func (m *PersistenceModule) shouldHydrateGenesisDb() (bool, error) {
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
