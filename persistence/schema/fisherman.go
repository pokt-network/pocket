package schema

var _ ProtocolActorSchema = &FishermanSchema{}

type FishermanSchema struct {
	BaseProtocolActorSchema
}

var FishermanActor ProtocolActorSchema = &FishermanSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		tableName:       "fisherman",
		chainsTableName: "fisherman_chains",

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       "fisherman_height",
		chainsHeightConstraintName: "fisherman_chain_height",
	},
}
