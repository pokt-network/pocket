package schema

// TODO (Team) NOTE: omitting 'missed blocks' for fear of creating a new record every time a validator misses a block
// TODO - likely will use block store and *byzantine validators* to process

const (
	ValTableName   = "val"
	ValTableSchema = `(
			id         BIGINT PRIMARY KEY,
			address    STRING NOT NULL UNIQUE, /*look into this being a "computed" field*/
			public_key STRING NOT NULL UNIQUE
		)`
	ValMetaTableName   = "val_meta"
	ValMetaTableSchema = `(
			val_id           BIGINT,
			height           BIGINT NOT NULL,
			service_url      STRING NOT NULL,
			staked_tokens    STRING NOT NULL,
			output_address   STRING  NOT NULL,
			paused           BOOL   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,
			constraint val_id_key_fk foreign key (val_id) references val (id)
		)`
)
