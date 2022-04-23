package schema

const (
	AccountTableName   = "Account"
	AccountTableSchema = `(
			id         BIGINT PRIMARY KEY,
			address    STRING NOT NULL UNIQUE,
		)`
	AccountMetaTableName   = "Account_meta"
	AccountMetaTableSchema = `(
			account_id        BIGINT,
			balance           STRING,
			constraint account_id_key_fk foreign key (account_id) references Account (id)
		)`

	PoolTableName   = "Pool"
	PoolTableSchema = `(
			id         BIGINT PRIMARY KEY,
			name       STRING NOT NULL UNIQUE,
			address    STRING NOT NULL UNIQUE,
		)`
	PoolMetaTableName   = "Pool_meta"
	PoolMetaTableSchema = `(
			pool_id        BIGINT,
			balance        STRING,
			constraint pool_id_key_fk foreign key (pool_id) references Pool (id)
		)`
)
