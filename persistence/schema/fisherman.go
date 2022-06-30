package schema

var _ ProtocolActorSchema = &FishermanSchema{}

type FishermanSchema struct {
	GenericProtocolActor
}

var FishermanActor ProtocolActorSchema = &FishermanSchema{
	GenericProtocolActor: GenericProtocolActor{
		tableName:       "fisherman",
		chainsTableName: "fisherman_chains",

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       "fisherman_height",
		chainsHeightConstraintName: "fisherman_chain_height",
	},
}
