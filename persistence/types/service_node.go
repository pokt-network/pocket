package types

var _ ProtocolActorSchema = &ServiceNodeSchema{}

type ServiceNodeSchema struct {
	BaseProtocolActorSchema
}

const (
	ServiceNodeTableName            = "service_node"
	ServiceNodeChainsTableName      = "service_node_chains"
	ServiceNodeHeightConstraintName = "service_node_height"
	ServiceNodeChainsConstraintName = "service_node_chain_height"
)

var ServiceNodeActor ProtocolActorSchema = &ServiceNodeSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		tableName:       ServiceNodeTableName,
		chainsTableName: ServiceNodeChainsTableName,

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       ServiceNodeHeightConstraintName,
		chainsHeightConstraintName: ServiceNodeChainsConstraintName,
	},
}
