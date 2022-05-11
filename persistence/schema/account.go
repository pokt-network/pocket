package schema

import "fmt"

const (
	AccountTableName   = "account"
	AccountTableSchema = `(
			address    TEXT NOT NULL,
			balance    TEXT NOT NULL,
			height 	   BIGINT NOT NULL,
			end_height BIGINT NOT NULL default -1
		)`
	AccountUniqueCreateIndex = `CREATE UNIQUE INDEX IF NOT EXISTS account_create_height ON account (address, height)`
	AccountUniqueDeleteIndex = `CREATE UNIQUE INDEX IF NOT EXISTS account_end_height ON account (address, end_height)`
)

func GetAccountAmountQuery(address string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE address='%s' AND end_height=%d`,
		AccountTableName, address, DefaultEndHeight)
}

func InsertAccountAmountQuery(address, amount string, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s (address, balance, height, end_height)
			VALUES ('%s','%s',%d, %d)
			ON CONFLICT (address, height)
			DO UPDATE SET balance='%s', end_height=%d
		`, AccountTableName, address, amount, height, DefaultEndHeight, amount, DefaultEndHeight)
}

func NullifyAccountAmountQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s'AND end_height=%d`,
		AccountTableName, height, address, DefaultEndHeight)
}

const (
	PoolTableName   = "pool"
	PoolTableSchema = `(
		name       TEXT NOT NULL,
		balance    TEXT NOT NULL,
		height 	   BIGINT NOT NULL,
		end_height BIGINT NOT NULL default -1
	)`
	PoolUniqueCreateIndex = `CREATE UNIQUE INDEX IF NOT EXISTS pool_create_height ON pool (name, height)`
	PoolUniqueDeleteIndex = `CREATE UNIQUE INDEX IF NOT EXISTS pool_end_height ON pool (name, end_height)`
)

func GetPoolAmountQuery(name string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE name='%s' AND end_height=%d`,
		PoolTableName, name, DefaultEndHeight)
}

func InsertPoolAmountQuery(name, amount string, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s (name, balance, height, end_height)
			VALUES ('%s','%s',%d, %d)
			ON CONFLICT (name, height)
			DO UPDATE SET balance='%s', end_height=%d
		`, PoolTableName, name, amount, height, DefaultEndHeight, amount, DefaultEndHeight)
}

func NullifyPoolAmountQuery(name string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE name='%s' AND end_height=%d`,
		PoolTableName, height, name, DefaultEndHeight)
}
