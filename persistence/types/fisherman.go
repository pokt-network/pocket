package types

import coreTypes "github.com/pokt-network/pocket/shared/core/types"

var _ ProtocolActorSchema = &FishermanSchema{}

type FishermanSchema struct {
	BaseProtocolActorSchema
}

const (
	FishermanTableName            = "fisherman"
	FishermanChainsTableName      = "fisherman_chains"
	FishermanHeightConstraintName = "fisherman_height"
	FishermanChainsConstraintName = "fisherman_chain_height"
)

var FishermanActor ProtocolActorSchema = &FishermanSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		actorType: coreTypes.ActorType_ACTOR_TYPE_FISH,

		tableName:       FishermanTableName,
		chainsTableName: FishermanChainsTableName,

		actorSpecificColName: ServiceUrlCol,

		heightConstraintName:       FishermanHeightConstraintName,
		chainsHeightConstraintName: FishermanChainsConstraintName,
	},
}
