package types

import "fmt"

// TODO(olshansky/team): Compare with `block.proto` and expand on this.
const (
	BlockTableName   = "block"
	BlockTableSchema = `(
			height             BIGINT PRIMARY KEY,
			hash 	           TEXT NOT NULL,
			proposer_address   BYTEA NOT NULL,
			quorum_certificate BYTEA NOT NULL
		)`
)

func InsertBlockQuery(height uint64, hashString string, proposerAddr, quorumCert []byte) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, hash, proposer_address, quorum_certificate)
			VALUES(%d, '%s', '%b', '%b')`,
		BlockTableName,
		height, hashString, proposerAddr, quorumCert)
}

func GetBlockHashQuery(height int64) string {
	return fmt.Sprintf(`SELECT hash FROM %s WHERE height=%d`, BlockTableName, height)
}

func GetMaximumBlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MAX(height) FROM %s`, BlockTableName)
}

func GetMinimumlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MIN(height) FROM %s`, BlockTableName)
}

func ClearAllBlocksQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, BlockTableName)
}
