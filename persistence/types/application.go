package types

import coreTypes "github.com/pokt-network/pocket/shared/core/types"

// TECHDEBT: s/ApplicationSchema/applicationSchema
// TECHDEBT: s/ApplicationActor/ApplicationActorSchema
// TECHDEBT: Ditto for all other actor types
// TECHDEBT: Should we use an ORM?
var _ ProtocolActorSchema = &ApplicationSchema{}

type ApplicationSchema struct {
	BaseProtocolActorSchema
}

const (
	AppTableName            = "app"
	AppChainsTableName      = "app_chains"
	AppHeightConstraintName = "app_height"
	AppChainsConstraintName = "app_chain_height"
)

var ApplicationActor ProtocolActorSchema = &ApplicationSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		actorType: coreTypes.ActorType_ACTOR_TYPE_APP,

		tableName:       AppTableName,
		chainsTableName: AppChainsTableName,

		actorSpecificColName: UnusedCol,

		heightConstraintName:       AppHeightConstraintName,
		chainsHeightConstraintName: AppChainsConstraintName,
	},
}
