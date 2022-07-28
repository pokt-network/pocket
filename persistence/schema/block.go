package schema

import "fmt"

// TODO(olshansky/team): Compare with `block.proto` and expand on this.
const (
	BlockTableName   = "block"
	BlockTableSchema = `(
			height             BIGINT PRIMARY KEY,
			hash 	           TEXT NOT NULL,
			proposer_address   TEXT NOT NULL,
			quorum_certificate BYTEA NOT NULL,
			transactions       BYTEA NOT NULL
		)`
)

func InsertBlockQuery(height uint64, hash string, proposerAddr []byte, quorumCert []byte, transactions [][]byte) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, hash, proposer_address, quorum_certificate, transactions)
			VALUES(%d, '%s', '%s', '%s', '%s')`,
		BlockTableName,
		height, hash, proposerAddr, quorumCert, transactions)
}

func GetBlockHashQuery(height int64) string {
	return fmt.Sprintf(`SELECT hash FROM %s WHERE height=%d`, BlockTableName, height)
}

func GetLatestBlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MAX(height) FROM %s`, BlockTableName)
}
