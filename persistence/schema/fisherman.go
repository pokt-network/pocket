package schema

type FishermanSchema struct {
	GenericProtocolActor
}

var FishermanActor ProtocolActor = &FishermanSchema{
	GenericProtocolActor: GenericProtocolActor{
		tableName:       "fisherman",
		chainsTableName: "fisherman_chains",

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       "fish_height",
		chainsHeightConstraintName: "fish_chain_height",
	},
}
