package schema

const (
	FishTableName   = "fisherman"
	FishTableSchema = `(
			id         UUID PRIMARY KEY,
			address    TEXT NOT NULL UNIQUE, /*TODO(andrew): look into this being a "computed" field*/
			public_key TEXT NOT NULL UNIQUE
		)`
	FishMetaTableName   = "fisherman_meta"
	FishMetaTableSchema = `(
			fisherman_id          UUID   NOT NULL,
			height           BIGINT NOT NULL,
			service_url      TEXT   NOT NULL,
			staked_tokens    TEXT   NOT NULL,
			output_address   TEXT   NOT NULL,
			paused           BOOL   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,

			constraint fisherman_id_key_fk foreign key (fisherman_id) references fisherman (id)
		)`
	FishChainsTableName   = "fisherman_chains"
	FishChainsTableSchema = `(
			fisherman_id      UUID    NOT NULL,
			chain_id     CHAR(4),
			height_start BIGINT  NOT NULL,
			height_end   BIGINT  NOT NULL default -1,

			constraint fisherman_id_key_fk foreign key (fisherman_id) references fisherman (id)
		)`
)
