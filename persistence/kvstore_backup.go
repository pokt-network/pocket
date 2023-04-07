package persistence

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
)

// GetBackupableKVStores returns the node and value kv stores allowing the ability to backup them
func (p *PostgresContext) GetBackupableKVStores() (nodeStores, valueStores map[int]kvstore.BackupableKVStore) {
	return p.stateTrees.GetBackupableKVStores()
}
