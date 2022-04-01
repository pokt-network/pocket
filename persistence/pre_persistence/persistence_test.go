package pre_persistence

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

func NewTestingPrePersistenceModule(_ *testing.T) *PrePersistenceModule {
	db := memdb.New(comparer.DefaultComparer, 10000000)
	cfg := &config.Config{Genesis: genesisJson()}
	_ = typesGenesis.GetNodeState(cfg)
	return NewPrePersistenceModule(db, types.NewMempool(10000, 10000), cfg)
}

func NewTestingPrePersistenceContext(t *testing.T) modules.PersistenceContext {
	persistenceModule := NewTestingPrePersistenceModule(t)
	persistenceContext, err := persistenceModule.NewContext(0)
	if err != nil {
		t.Fatal(err)
	}
	return persistenceContext
}

func genesisJson() string {
	return fmt.Sprintf(`{
		"genesis_state_configs": {
			"num_validators": 5,
			"num_applications": 1,
			"num_fisherman": 1,
			"num_servicers": 5,
			"keys_seed_start": %d
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash"
	}`, 42)
}
