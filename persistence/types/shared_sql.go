package types

import "fmt"

func SelectJSON(selectStatement string) string {
	return fmt.Sprintf(`
	SELECT json_agg(row_to_json(subquery.*))
	FROM (
		%s
	) AS subquery;`, selectStatement)
}
