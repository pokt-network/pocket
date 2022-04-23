package schema

// TODO (Team) NOTE: omitting 'missed blocks' for fear of creating a new record every time a validator misses a block
// TODO - likely will use block store and *byzantine validators* to process

const (
	ValTableName   = "validator"
	ValTableSchema = `(
			id         UUID PRIMARY KEY,
			address    TEXT NOT NULL UNIQUE, /*look into this being a "computed" field*/
			public_key TEXT NOT NULL UNIQUE
		)`
	ValMetaTableName   = "validator_meta"
	ValMetaTableSchema = `(
			val_id           UUID NOT NULL,
			height           BIGINT NOT NULL,
			service_url      TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			output_address   TEXT  NOT NULL,
			paused           BOOL   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,
			constraint val_id_key_fk foreign key (val_id) references val (id)
		)`
)
