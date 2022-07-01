package schema

var _ ProtocolActorSchema = &ApplicationSchema{}

type ApplicationSchema struct {
	BaseProtocolActorSchema
}

var ApplicationActor ProtocolActorSchema = &ApplicationSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		tableName:       "app",
		chainsTableName: "app_chains",

		actorSpecificColName: MaxRelaysCol,

		heightConstraintName:       "app_height",
		chainsHeightConstraintName: "app_chain_height",
	},
}
