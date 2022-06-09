package schema

import "fmt"

const (
	AccountTableName   = "account"
	AccountTableSchema = `(
			address    TEXT NOT NULL,
			balance    TEXT NOT NULL,
			height 	   BIGINT NOT NULL,
		    CONSTRAINT account_create_height UNIQUE (address, height)
		)`
)

func GetAccountAmountQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE address='%s' AND height=%d ORDER BY height DESC LIMIT 1`,
		AccountTableName, address, height)
}

func InsertAccountAmountQuery(address, amount string, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s (address, balance, height)
			VALUES ('%s','%s',%d)
			ON CONFLICT ON CONSTRAINT account_create_height
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height
		`, AccountTableName, address, amount, height)
}

const (
	PoolTableName   = "pool"
	PoolTableSchema = `(
		name       TEXT NOT NULL,
		balance    TEXT NOT NULL,
		height 	   BIGINT NOT NULL,

		CONSTRAINT pool_create_height UNIQUE (name, height)
	)`
)

func GetPoolAmountQuery(name string, height int64) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE name='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		PoolTableName, name, height)
}

func InsertPoolAmountQuery(name, amount string, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s (name, balance, height)
			VALUES ('%s','%s',%d)
			ON CONFLICT ON CONSTRAINT pool_create_height
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height
		`, PoolTableName, name, amount, height)
}
