package persistence

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.PersistenceModule = &persistenceModule{}
var _ modules.PersistenceRWContext = &PostgresContext{}

type persistenceModule struct {
	postgresURL string
	nodeSchema  string
	db          *pgx.Conn
	bus         modules.Bus
}

func NewPersistenceModule(postgresURL string, nodeSchema string, db *pgx.Conn, bus modules.Bus) *persistenceModule {
	return &persistenceModule{postgresURL: postgresURL, nodeSchema: nodeSchema, db: db, bus: bus}
}

func Create(c *config.Config) (modules.PersistenceModule, error) {
	db, err := ConnectAndInitializeDatabase(c.Persistence.PostgresUrl, c.Persistence.NodeSchema)
	if err != nil {
		return nil, err
	}
	pm := NewPersistenceModule(c.Persistence.PostgresUrl, c.Persistence.NodeSchema, db, nil)
	// populate genesis state
	pm.PopulateGenesisState(c.GenesisSource.GetState())
	return pm, nil
}

func (p *persistenceModule) Start() error {
	log.Println("Starting persistence module...")
	return nil
}

func (p *persistenceModule) Stop() error {
	log.Println("Stopping persistence module...")
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
	db, err := ConnectAndInitializeDatabase(m.postgresURL, m.nodeSchema)
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
		DB:     PostgresDB{tx},
	}, nil
}

func (m *persistenceModule) NewReadContext(height int64) (modules.PersistenceReadContext, error) {
	return m.NewRWContext(height)
	// TODO (Team) this can be completely separate from rw context.
	// It should access the db directly rather than using transactions
}
