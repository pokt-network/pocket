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

func ClearAllIBCQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, IBCStoreTableName)
}
