package ibc

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var (
	_ modules.StoreManager  = (*Stores)(nil)
	_ modules.ProvableStore = (*ProvableStore)(nil)
	_ modules.Store         = (*PrivateStore)(nil)
)

type Stores struct {
	stores map[string]modules.Store
}

type ProvableStore struct {
	nodeStore kvstore.KVStore
	tree      *smt.SMT
	storeKey  string
}

type PrivateStore struct {
	nodeStore kvstore.KVStore
}
