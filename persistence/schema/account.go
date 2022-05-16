package schema

import "fmt"

const (
	AccountTableName   = "account"
	AccountTableSchema = `(
			address    TEXT NOT NULL,
			balance    TEXT NOT NULL,
			height 	   BIGINT NOT NULL,
			end_height BIGINT NOT NULL default -1,

			/* DISCUSS(drewsky): We can't do ON CONFLICT multiple constraints, so what should we do here? */
			CONSTRAINT account_create_height UNIQUE (address, height),
			CONSTRAINT account_end_height UNIQUE (address, end_height)
		)`
)

func GetAccountAmountQuery(address string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE address='%s' AND end_height=%d`,
		AccountTableName, address, DefaultEndHeight)
}

func InsertAccountAmountQuery(address, amount string, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s (address, balance, height, end_height)
			VALUES ('%s','%s',%d, %d)
			ON CONFLICT ON CONSTRAINT account_create_height
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height, end_height=EXCLUDED.end_height
		`, AccountTableName, address, amount, height, DefaultEndHeight)
}

func NullifyAccountAmountQuery(address string, height int64) string {
	return fmt.Sprintf(`
		UPDATE %s SET end_height=%d WHERE address='%s'AND end_height=%d`,
		AccountTableName, height, address, DefaultEndHeight)
}

const (
	PoolTableName   = "pool"
	PoolTableSchema = `(
		name       TEXT NOT NULL,
		balance    TEXT NOT NULL,
		height 	   BIGINT NOT NULL,
		end_height BIGINT NOT NULL default -1,

		CONSTRAINT pool_create_height UNIQUE (name, height),
		CONSTRAINT pool_end_height UNIQUE (name, end_height)
	)`
)

func GetPoolAmountQuery(name string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE name='%s' AND end_height=%d`,
		PoolTableName, name, DefaultEndHeight)
}

func InsertPoolAmountQuery(name, amount string, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s (name, balance, height, end_height)
			VALUES ('%s','%s',%d, %d)
			ON CONFLICT ON CONSTRAINT pool_create_height
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height, end_height=EXCLUDED.end_height
		`, PoolTableName, name, amount, height, DefaultEndHeight)
}

func NullifyPoolAmountQuery(name string, height int64) string {
	return fmt.Sprintf(`
		UPDATE %s SET end_height=%d WHERE name='%s' AND end_height=%d`,
		PoolTableName, height, name, DefaultEndHeight)
}
