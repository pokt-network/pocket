package schema

import "fmt"

// TODO(olshansky/team): Compare with `block.proto` and expand on this.
const (
	BlockTableName   = "block"
	BlockTableSchema = `(
			height             BIGINT PRIMARY KEY,
			hash 	           TEXT NOT NULL,
			proposer_address   TEXT NOT NULL,
			quorum_certificate TEXT NOT NULL
		)`
)

func InsertBlockQuery(height uint64, hashString string, proposerAddr []byte, quorumCert []byte) string {
	return fmt.Sprintf(
		`INSERT INTO %s(height, hash, proposer_address, quorum_certificate)
			VALUES(%d, '%s', '%s', '%s')`,
		BlockTableName,
		height, hashString, proposerAddr, quorumCert)
}

func GetBlockHashQuery(height int64) string {
	return fmt.Sprintf(`SELECT hash FROM %s WHERE height=%d`, BlockTableName, height)
}

func GetLatestBlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MAX(height) FROM %s`, BlockTableName)
}
