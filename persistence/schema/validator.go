package schema

var _ ProtocolActor = &ValidatorSchema{}

type ValidatorSchema struct {
	GenericProtocolActor
}

var ValidatorActor ProtocolActor = &ServiceNodeSchema{
	GenericProtocolActor: GenericProtocolActor{
		tableName: "validator",

		actorSpecificColName: ServiceURLCol,

		heightConstraintName: "validator_node_height",
	},
}
