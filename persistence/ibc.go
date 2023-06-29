package persistence

import (
	pTypes "github.com/pokt-network/pocket/persistence/types"
)

// SetIBCStoreEntry sets the key value pair in the IBC store postgres table
func (p *PostgresContext) SetIBCStoreEntry(key, value []byte) error {
	ctx, tx := p.getCtxAndTx()
	if _, err := tx.Exec(ctx, pTypes.InsertIBCStoreEntryQuery(p.Height, key, value)); err != nil {
		return err
	}
	return nil
}
