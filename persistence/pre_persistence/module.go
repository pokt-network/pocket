package pre_persistence

import (
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

func Create(cfg *config.Config) (modules.PersistenceModule, error) {
	db := memdb.New(comparer.DefaultComparer, 888888888)
	state := GetTestState()
	state.LoadStateFromConfig(cfg)
	return NewPrePersistenceModule(db, types.NewMempool(10000, 10000), cfg), nil

}

func (p *PrePersistenceModule) Start() error {
	pCtx, err := p.NewContext(0)
	if err != nil {
		return err
	}
	genesis, _, _, _, _, err := NewGenesisState(p.Cfg, 5, 1, 1, 5)
	if err != nil {
		return err
	}
	c := pCtx.(*PrePersistenceContext)
	err = InitGenesis(c, genesis)
	if err != nil {
		return err
	}
	err = c.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (p *PrePersistenceModule) Stop() error {
	return nil
}

//func (m *PrePersistenceModule) SetBus(pocketBus modules.Bus) { TODO: add back in once bus is integrated
//	m.pocketBusMod = pocketBus
//}
//
//func (m *PrePersistenceModule) GetBus() modules.Bus {
//	if m.pocketBusMod == nil {
//		log.Fatalf("PocketBus is not initialized")
//	}
//	return m.pocketBusMod
//}
