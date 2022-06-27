package schema

var _ ProtocolActor = &ApplicationSchema{}

type ApplicationSchema struct {
	GenericProtocolActor
}

var ApplicationActor ProtocolActor = &ApplicationSchema{
	GenericProtocolActor: GenericProtocolActor{
		tableName:       "app",
		chainsTableName: "app_chains",

		actorSpecificColName: MaxRelaysCol,

		heightConstraintName:       "app_height",
		chainsHeightConstraintName: "app_chain_height",
	},
}
