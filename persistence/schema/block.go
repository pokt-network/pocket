package schema

import "fmt"

const (
	// TODO(olshansk): needs to be implemented (and tests obviously), only height is here because it's the MVP requirement for utility
	BlockTableName   = "block"
	BlockTableSchema = `(
			height  BIGINT PRIMARY KEY,
			hash 	TEXT NOT NULL
		)`
)

func BlockHashQuery(height int64) string {
	return fmt.Sprintf(`SELECT hash FROM %s WHERE height=%d`, BlockTableName, height)
}

func LatestBlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MAX(height) FROM %s`, BlockTableName)
}
