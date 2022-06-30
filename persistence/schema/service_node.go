package schema

var _ ProtocolActorSchema = &ServiceNodeSchema{}

type ServiceNodeSchema struct {
	GenericProtocolActor
}

var ServiceNodeActor ProtocolActorSchema = &ServiceNodeSchema{
	GenericProtocolActor: GenericProtocolActor{
		tableName:       "service_node",
		chainsTableName: "service_node_chains",

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       "service_node_height",
		chainsHeightConstraintName: "service_node_chain_height",
	},
}
