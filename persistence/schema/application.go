package schema

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
		tableName:       AppTableName,
		chainsTableName: AppChainsTableName,

		actorSpecificColName: MaxRelaysCol,

		heightConstraintName:       AppHeightConstraintName,
		chainsHeightConstraintName: AppChainsConstraintName,
	},
}
