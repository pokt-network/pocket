package persistence

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	pTypes "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// SetIBCStoreEntry sets the key value pair in the IBC store postgres table at the current height
func (p *PostgresContext) SetIBCStoreEntry(key, value []byte) error {
	ctx, tx := p.getCtxAndTx()
	if _, err := tx.Exec(ctx, pTypes.InsertIBCStoreEntryQuery(uint64(p.Height), key, value)); err != nil {
		return err
	}
	return nil
}

// GetIBCStoreEntry returns the stored value for the key at the height provided from the IBC store table
func (p *PostgresContext) GetIBCStoreEntry(key []byte, height uint64) ([]byte, error) {
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

// SetIBCEvent sets the IBC event at the current height in the persitence DB
func (p *PostgresContext) SetIBCEvent(event *coreTypes.IBCEvent) error {
	ctx, tx := p.getCtxAndTx()
	typeStr := event.GetTopic()
	eventBz, err := codec.GetCodec().Marshal(event)
	if err != nil {
		return err
	}
	eventHex := hex.EncodeToString(eventBz)
	if _, err := tx.Exec(ctx, pTypes.InsertIBCEventQuery(uint64(p.Height), typeStr, eventHex)); err != nil {
		return err
	}
	return nil
}

// GetIBCEvents returns all the IBC events at the height provided with the matching topic
func (p *PostgresContext) GetIBCEvents(height uint64, topic string) ([]*coreTypes.IBCEvent, error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, pTypes.GetIBCEventQuery(height, topic))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []*coreTypes.IBCEvent
	for rows.Next() {
		var eventHex string
		if err := rows.Scan(&eventHex); err != nil {
			return nil, err
		}
		eventBz, err := hex.DecodeString(eventHex)
		if err != nil {
			return nil, err
		}
		event := &coreTypes.IBCEvent{}
		if err := codec.GetCodec().Unmarshal(eventBz, event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
