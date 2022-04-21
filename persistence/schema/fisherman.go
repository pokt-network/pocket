package schema

const (
	fishTableName   = "fish"
	fishTableSchema = `(
			id         BIGINT PRIMARY KEY,
			address    STRING NOT NULL UNIQUE, /*look into this being a "computed" field*/
			public_key STRING NOT NULL UNIQUE
		)`
	fishMetaTableName   = "fish_meta"
	fishMetaTableSchema = `(
			fish_id           BIGINT,
			height           BIGINT NOT NULL,
			service_url      STRING NOT NULL,
			staked_tokens    STRING NOT NULL,
			output_address   STRING  NOT NULL,
			paused           bool   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,
			constraint fish_id_key_fk foreign key (fish_id) references fish (id)
		)`
	fishChainsTableName  = "fish_chains"
	fishChainTableSchema = `(
			fish_id       BIGINT,
			chainId      CHAR(4),
			height_start BIGINT NOT NULL,
			height_end   BIGINT NOT NULL default -1,
			constraint fish_id_key_fk foreign key (fish_id) references fish (id)
		)`
)
