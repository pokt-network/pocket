package persistence

import (
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"pocket/persistence/pre_persistence"
	"pocket/shared/config"
	"pocket/shared/modules"
	types2 "pocket/utility/types"
)

type persistenceModule struct {
	modules.PersistenceModule

	CommitDB     *memdb.DB
	Mempool      types2.Mempool
	pocketBusMod modules.Bus
	Cfg          *config.Config
}

func Create(cfg *config.Config) (modules.PersistenceModule, error) {
	db := memdb.New(comparer.DefaultComparer, 888888888)
	state := pre_persistence.GetTestState()
	state.LoadStateFromConfig(cfg)
	return &persistenceModule{
		PersistenceModule: nil,
		CommitDB:          db,
		Cfg:               cfg,
		Mempool:           types2.NewMempool(1000, 1000),
		pocketBusMod:      nil,
	}, nil

}

func (p *persistenceModule) Start() error {
	pCtx, err := p.NewContext(0)
	if err != nil {
		return err
	}
	genesis, _, _, _, _, err := pre_persistence.NewMockGenesisState(5, 1, 1, 5)
	if err != nil {
		return err
	}
	c := pCtx.(*pre_persistence.MockPersistenceContext)
	err = pre_persistence.InitGenesis(c, genesis)
	if err != nil {
		return err
	}
	err = c.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (p *persistenceModule) Stop() error {
	return nil
}

func (m *persistenceModule) SetBus(pocketBus modules.Bus) {
	m.pocketBusMod = pocketBus
}

func (m *persistenceModule) GetBus() modules.Bus {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *persistenceModule) NewContext(height int64) (modules.PersistenceContext, error) {
	newDB := pre_persistence.NewMemDB()
	it := m.CommitDB.NewIterator(&util.Range{Start: pre_persistence.HeightKey(height, nil), Limit: pre_persistence.HeightKey(height+1, nil)})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		err := newDB.Put(pre_persistence.KeyFromHeightKey(it.Key()), it.Value())
		if err != nil {
			return nil, err
		}
	}
	context := &pre_persistence.MockPersistenceContext{
		Height: height,
		Parent: m,
		//SavePoints: make(map[string]int),
		DBs: make([]*memdb.DB, 0),
	}
	//context.SavePoints[hex.EncodeToString(pre_persistence.FirstSavePointKey)] = types2.ZeroInt
	context.DBs = append(context.DBs, newDB)
	return context, nil
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	return m.CommitDB
}
