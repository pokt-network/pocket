package schema

const (
	ServiceNodeTableName   = "service_node"
	ServiceNodeTableSchema = `(
			id         UUID PRIMARY KEY,
			address    TEXT NOT NULL UNIQUE, /*TODO(andrew): look into this being a "computed" field*/
			public_key TEXT NOT NULL UNIQUE
		)`
	ServiceNodeMetaTableName   = "service_node_meta"
	ServiceNodeMetaTableSchema = `(
			service_node_id  UUID NOT NULL,
			height           BIGINT not null,
			service_url      TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			output_address   TEXT  NOT NULL,
			paused           BOOL   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,
			constraint service_node_id_key_fk foreign key (service_node_id) references service_node (id)
		)`
	ServiceNodeChainsTableName   = "service_node_chains"
	ServiceNodeChainsTableSchema = `(
			service_node_id   UUID NOT NULL,
			chain_id          CHAR(4),
			height_start      BIGINT NOT NULL,
			height_end        BIGINT NOT NULL default -1,
			constraint service_node_id_key_fk foreign key (service_node_id) references service_node (id)
		)`
)
