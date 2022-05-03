package schema

import "fmt"

const (
	// TODO (OLSHANK) needs to be implemented, only height is here because it's the mvp requirement for utility
	BlockTableName   = "block"
	BlockTableSchema = `(
			Height  BIGINT PRIMARY KEY,
			Hash 	TEXT NOT NULL
		)`
)

func BlockHashQuery(height int64) string {
	return fmt.Sprintf(`SELECT hash FROM %s WHERE Height=%d`, BlockTableName, height)
}

func LatestBlockHeightQuery() string {
	return fmt.Sprintf(`SELECT MAX(Height) FROM %s`, BlockTableName)
}
