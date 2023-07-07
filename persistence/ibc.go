package persistence

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	pTypes "github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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
	var valueHex string
	err := row.Scan(&valueHex)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, coreTypes.ErrIBCKeyDoesNotExist(string(key))
	} else if err != nil {
		return nil, err
	}
	value, err := hex.DecodeString(valueHex)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(value, nil) {
		return nil, coreTypes.ErrIBCKeyDoesNotExist(string(key))
	}
	return value, nil
}
