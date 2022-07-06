package schema

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
		tableName:       FishermanTableName,
		chainsTableName: FishermanChainsTableName,

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       FishermanHeightConstraintName,
		chainsHeightConstraintName: FishermanChainsConstraintName,
	},
}
