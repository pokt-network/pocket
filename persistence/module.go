package persistence

import (
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"log"
	"pocket/consensus/pkg/config"
	"pocket/shared/context"
	"pocket/shared/modules"
	"pocket/utility/utility/types"
)

type persistenceModule struct {
	modules.PersistenceModule

	memdb        *memdb.DB
	Mempool      types.Mempool
	pocketBusMod modules.PocketBusModule
}

func Create(config *config.Config) (modules.PersistenceModule, error) {

	db := memdb.New(comparer.DefaultComparer, 888888888)

	return &persistenceModule{
		Mempool: types.NewMempool(1000, 1000),
		memdb: db,
	}, nil

}

func(p *persistenceModule) Start(ctx *context.PocketContext) error {
	return nil
}

func(p *persistenceModule) Stop(*context.PocketContext) error {
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