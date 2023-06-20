package types

import (
	"encoding/hex"
	"fmt"
)

const (
	IbcStoreTableName   = "ibc_messages"
	IbcStoreTableSchema = `(
		height BIGINT NOT NULL,
		key TEXT NOT NULL,
		value TEXT,
		PRIMARY KEY (height, key)
	)`
)

func InsertIbcStoreEntryQuery(height int64, key, value []byte) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, key, value) VALUES(%d, '%s', '%s')`,
		IbcStoreTableName,
		height,
		hex.EncodeToString(key),
		hex.EncodeToString(value),
	)
}
