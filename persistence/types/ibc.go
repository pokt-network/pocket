package types

import (
	"encoding/hex"
	"fmt"
)

const (
	IBCStoreTableName   = "ibc_entries"
	IBCStoreTableSchema = `(
		height BIGINT NOT NULL,
		key TEXT NOT NULL,
		value TEXT,
		PRIMARY KEY (height, key)
	)`
)

func InsertIBCStoreEntryQuery(height int64, key, value []byte) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, key, value) VALUES(%d, '%s', '%s')`,
		IBCStoreTableName,
		height,
		hex.EncodeToString(key),
		hex.EncodeToString(value),
	)
}

// Return the latest value for the key at the height provided or at the last updated height
func GetIBCStoreEntryQuery(height int64, key []byte) string {
	return fmt.Sprintf(
		`SELECT value FROM %s WHERE height <= %d AND key = '%s' ORDER BY height DESC LIMIT 1`,
		IBCStoreTableName,
		height,
		hex.EncodeToString(key),
	)
}

func ClearAllIBCQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, IBCStoreTableName)
}
