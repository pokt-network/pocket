package schema

const (
	AppTableName   = "app"
	AppTableSchema = `(
			id         UUID PRIMARY KEY,
			address    BYTEA NOT NULL UNIQUE, /*TODO(andrew): look into this being a "computed" field*/
			public_key BYTEA NOT NULL UNIQUE
		)`
	AppMetaTableName   = "app_meta"
	AppMetaTableSchema = `(
			app_id           UUID   NOT NULL,
			height           BIGINT NOT NULL,
			staked_tokens    TEXT   NOT NULL,
			max_relays       TEXT   NOT NULL,  /*TODO(andrew): look into this being a "computed" field*/
			output_address   BYTEA  NOT NULL,
			paused           BOOL   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,

			constraint app_id_key_fk foreign key (app_id) references app (id)
		)`
	AppChainsTableName   = "app_chains"
	AppChainsTableSchema = `(
			app_id       UUID NOT NULL,
			chain_id     CHAR(4),
			height_start BIGINT NOT NULL,
			height_end   BIGINT NOT NULL default -1,

			constraint app_id_key_fk foreign key (app_id) references app (id)
		)`
)
