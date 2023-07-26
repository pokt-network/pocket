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
		value TEXT NOT NULL,
		PRIMARY KEY (height, key)
	)`
	IBCEventLogTableName   = "ibc_events"
	IBCEventLogTableSchema = `(
		height BIGINT NOT NULL,
		topic TEXT NOT NULL,
		event TEXT NOT NULL,
		PRIMARY KEY (height, topic, event)
	)`
)

// InsertIBCStoreEntryQuery returns the query to insert a key/value pair into the ibc_entries table
func InsertIBCStoreEntryQuery(height uint64, key, value []byte) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, key, value) VALUES(%d, '%s', '%s')`,
		IBCStoreTableName,
		height,
		hex.EncodeToString(key),
		hex.EncodeToString(value),
	)
}

// InsertIBCEventQuery returns the query to insert an event into the ibc_events table
func InsertIBCEventQuery(height uint64, topic, eventHex string) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, topic, event) VALUES(%d, '%s', '%s')`,
		IBCEventLogTableName,
		height,
		topic,
		eventHex,
	)
}

// GetIBCStoreEntryQuery returns the latest value for the key at the height provided or at the last updated height
func GetIBCStoreEntryQuery(height uint64, key []byte) string {
	return fmt.Sprintf(
		`SELECT value FROM %s WHERE height <= %d AND key = '%s' ORDER BY height DESC LIMIT 1`,
		IBCStoreTableName,
		height,
		hex.EncodeToString(key),
	)
}

// GetIBCEventQuery returns the query to get all events for a given height and topic
func GetIBCEventQuery(height uint64, topic string) string {
	return fmt.Sprintf(
		`SELECT event FROM %s WHERE height = %d AND topic = '%s'`,
		IBCEventLogTableName,
		height,
		topic,
	)
}

// ClearAllIBCStoreQuery returns the query to clear all entries from the ibc_entries table
func ClearAllIBCStoreQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, IBCStoreTableName)
}

// ClearAllIBCEventsQuery returns the query to clear all entries from the ibc_events table
func ClearAllIBCEventsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, IBCEventLogTableName)
}
