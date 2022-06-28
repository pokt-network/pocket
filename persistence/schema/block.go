package schema

import "fmt"

// TODO(olshansky/team): Implement this for consensus. Block height was only added for MVP purposes.
const (
	BlockTableName   = "block"
	BlockTableSchema = `(
			height  BIGINT PRIMARY KEY,
			hash 	TEXT NOT NULL
		)`
)

func GetBlockHashQuery(height int64) string {
	return fmt.Sprintf(`SELECT hash FROM %s WHERE height=%d`, BlockTableName, height)
}

func GetLatestBlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MAX(height) FROM %s`, BlockTableName)
}
