package schema

const (
	AppTableName   = "app"
	AppTableSchema = `(
			id         BIGINT PRIMARY KEY,
			address    BYTEA NOT NULL UNIQUE, /*look into this being a "computed" field*/
			public_key BYTEA NOT NULL UNIQUE
		)`
	AppMetaTableName   = "app_meta"
	AppMetaTableSchema = `(
			app_id           BIGINT,
			height           BIGINT NOT NULL,
			staked_tokens    STRING NOT NULL,
			max_relays		 STRING NOT NULL,  /*look into this being a "computed" field*/
			output_address   BYTEA  NOT NULL,
			paused           bool   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,
			constraint app_id_key_fk foreign key (app_id) references App (id)
		)`
	AppChainsTableName  = "app_chains"
	AppChainTableSchema = `(
			app_id       BIGINT,
			chainId      CHAR(4),
			height_start BIGINT NOT NULL,
			height_end   BIGINT NOT NULL default -1,
			constraint app_id_key_fk foreign key (app_id) references App (id)
		)`
)
