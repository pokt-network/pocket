package types

import "fmt"

func protocolAccountTableSchema(accountSpecificColName, constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s BIGINT NOT NULL,

		    CONSTRAINT %s UNIQUE (%s, %s)
		)`, accountSpecificColName, BalanceCol, HeightCol, constraintName, accountSpecificColName, HeightCol)
}

func SelectAccounts(height int64, colName, tableName string) string {
	return fmt.Sprintf(`
			SELECT DISTINCT ON (%s) %s, balance, height
			FROM %s
			WHERE height<=%d
			ORDER BY %s, height DESC
       `, colName, colName, tableName, height, colName)
}

func SelectBalance(accountSpecificParam, accountSpecificParamValue string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE %s='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		tableName, accountSpecificParam, accountSpecificParamValue, height)
}

func InsertAccount(accountSpecificParam, accountSpecificParamValue, amount string, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s (%s, balance, height)
			VALUES ('%s','%s',%d)
			ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height
		`, tableName, accountSpecificParam, accountSpecificParamValue, amount, height, constraintName)
}
