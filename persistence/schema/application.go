package schema

var _ ProtocolActorSchema = &ApplicationSchema{}

type ApplicationSchema struct {
	GenericProtocolActor
}

var ApplicationActor ProtocolActorSchema = &ApplicationSchema{
	GenericProtocolActor: GenericProtocolActor{
		tableName:       "app",
		chainsTableName: "app_chains",

		actorSpecificColName: MaxRelaysCol,

		heightConstraintName:       "app_height",
		chainsHeightConstraintName: "app_chain_height",
	},
}
