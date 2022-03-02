package pre_persistence

import (
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"testing"
)

func NewTestingPrePersistenceModule(_ *testing.T) *PrePersistenceModule {
	db := memdb.New(comparer.DefaultComparer, 10000000)
	return NewPrePersistenceModule(db, types.NewMempool(10000, 10000), &config.Config{})
}

func NewTestingPrePersistenceContext(t *testing.T) modules.PersistenceContext {
	persistenceModule := NewTestingPrePersistenceModule(t)
	persistenceContext, err := persistenceModule.NewContext(0)
	if err != nil {
		t.Fatal(err)
	}
	return persistenceContext
}
