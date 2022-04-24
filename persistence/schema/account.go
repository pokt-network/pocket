package schema

const (
	AccountTableName   = "account"
	AccountTableSchema = `(
			id      UUID PRIMARY KEY,
			address TEXT NOT NULL UNIQUE
		)`
	AccountMetaTableName   = "account_meta"
	AccountMetaTableSchema = `(
			account_id UUID,
			height     BIGINT,
			balance    TEXT,

			constraint account_id_key_fk foreign key (account_id) references account (id)
		)`

	PoolTableName   = "pool"
	PoolTableSchema = `(
			id      UUID PRIMARY KEY,
			name    TEXT NOT NULL UNIQUE,
			address TEXT NOT NULL UNIQUE
		)`
	PoolMetaTableName   = "pool_meta"
	PoolMetaTableSchema = `(
			pool_id UUID,
			height  BIGINT,
			balance TEXT,

			constraint pool_id_key_fk foreign key (pool_id) references pool (id)
		)`
)
