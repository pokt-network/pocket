package persistence

import (
	pTypes "github.com/pokt-network/pocket/persistence/types"
)

// SetIBCStoreEntry sets the key value pair in the IBC store postgres table at the current height
func (p *PostgresContext) SetIBCStoreEntry(key, value []byte) error {
	ctx, tx := p.getCtxAndTx()
	if _, err := tx.Exec(ctx, pTypes.InsertIBCStoreEntryQuery(p.Height, key, value)); err != nil {
		return err
	}
	return nil
}

// GetIBCStoreEntry returns the stored value for the key at the height provided from the IBC store table
func (p *PostgresContext) GetIBCStoreEntry(key []byte, height int64) ([]byte, error) {
	ctx, tx := p.getCtxAndTx()
	row := tx.QueryRow(ctx, pTypes.GetIBCStoreEntryQuery(height, key))
	var value []byte
	if err := row.Scan(&value); err != nil {
		return nil, err
	}
	return value, nil
}
