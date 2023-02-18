package types

import coreTypes "github.com/pokt-network/pocket/shared/core/types"

var _ ProtocolActorSchema = &ServicerSchema{}

type ServicerSchema struct {
	BaseProtocolActorSchema
}

const (
	ServicerTableName            = "servicer"
	ServicerChainsTableName      = "servicer_chains"
	ServicerHeightConstraintName = "servicer_height"
	ServicerChainsConstraintName = "servicer_chain_height"
)

var ServicerActor ProtocolActorSchema = &ServicerSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		actorType: coreTypes.ActorType_ACTOR_TYPE_SERVICER,

		tableName:       ServicerTableName,
		chainsTableName: ServicerChainsTableName,

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       ServicerHeightConstraintName,
		chainsHeightConstraintName: ServicerChainsConstraintName,
	},
}
