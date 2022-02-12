package persistence

import (
	"encoding/hex"
	"log"
	"pocket/consensus/pkg/config"
	"pocket/shared/context"
	"pocket/shared/modules"

	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/syndtr/goleveldb/leveldb/util"

	// REALLY BAD. WE SHOULD NEVER DO THIS.
	"pocket/utility/utility/test"
	"pocket/utility/utility/types"
)

type persistenceModule struct {
	modules.PersistenceModule

	CommitDB     *memdb.DB
	Mempool      types.Mempool
	pocketBusMod modules.PocketBusModule
}

func Create(config *config.Config) (modules.PersistenceModule, error) {
	db := memdb.New(comparer.DefaultComparer, 888888888)

	return &persistenceModule{
		Mempool:  types.NewMempool(1000, 1000),
		CommitDB: db,
	}, nil

}

func (p *persistenceModule) Start(ctx *context.PocketContext) error {
	return nil
}

func (p *persistenceModule) Stop(*context.PocketContext) error {
	return nil
}

func (m *persistenceModule) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *persistenceModule) GetPocketBusMod() modules.PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *persistenceModule) NewContext(height int64) (modules.PersistenceContext, error) {
	newDB := test.NewMemDB()
	it := m.CommitDB.NewIterator(&util.Range{Start: test.HeightKey(height, nil), Limit: test.HeightKey(height+1, nil)})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		err := newDB.Put(test.KeyFromHeightKey(it.Key()), it.Value())
		if err != nil {
			return nil, err
		}
	}
	context := &test.MockPersistenceContext{
		Height:     0,
		Parent:     m,
		SavePoints: make(map[string]int),
		DBs:        make([]*memdb.DB, 0),
	}
	context.SavePoints[hex.EncodeToString(test.FirstSavePointKey)] = types.ZeroInt
	context.DBs = append(context.DBs, newDB)
	return context, nil
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	return m.CommitDB
}
