package schema

const (
	serviceNodeTableName   = "service_node"
	serviceNodeTableSchema = `(
			id         bigint PRIMARY KEY,
			address    STRING NOT NULL UNIQUE, /*look into this being a "computed" field*/
			public_key STRING NOT NULL UNIQUE
		)`
	serviceNodeMetaTableName   = "service_node_meta"
	serviceNodeMetaTableSchema = `(
			service_node_id           bigint,
			height           BIGINT not null,
			service_url      STRING NOT NULL,
			staked_tokens    STRING NOT NULL,
			output_address   STRING  NOT NULL,
			paused           BOOL   NOT NULL default false,
			unstaking_height BIGINT NOT NULL default -1,
			constraint service_node_id_key_fk foreign key (service_node_id) references service_node (id)
		)`
	serviceNodeChainsTableName  = "service_node_chains"
	serviceNodeChainTableSchema = `(
			service_node_id       bigint,
			chainId      CHAR(4),
			height_start BIGINT NOT NULL,
			height_end   BIGINT NOT NULL default -1,
			constraint service_node_id_key_fk foreign key (service_node_id) references service_node (id)
		)`
)
