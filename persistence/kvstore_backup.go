package persistence

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
)

// GetKVStores returns the node and value kv stores allowing the ability to backup them
func (p *PostgresContext) GetKVStores() (nodeStores, valueStores map[int]kvstore.BackupableKVStore) {
	return p.stateTrees.GetKVStores()
}
