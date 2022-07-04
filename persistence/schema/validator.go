package schema

var _ ProtocolActorSchema = &ValidatorSchema{}

const (
	ValidatorTableName        = "validator"
	ValidatorHeightConstraint = "validator_node_height"
	ValPan                    = "not implemented for validator schema"
	NullString                = ""
)

type ValidatorSchema struct {
	BaseProtocolActorSchema
}

var ValidatorActor ProtocolActorSchema = &ValidatorSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		tableName:       ValidatorTableName,
		chainsTableName: NullString,

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       ValidatorHeightConstraint,
		chainsHeightConstraintName: NullString,
	},
}

func (actor *ValidatorSchema) InsertQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, _ []string, height int64) string {
	return Insert(BaseActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
	},
		actor.actorSpecificColName, maxRelays,
		actor.heightConstraintName, NullString,
		actor.tableName, NullString,
		height)
}

func (actor *ValidatorSchema) UpdateChainsQuery(_ string, _ []string, _ int64) string { panic(ValPan) }
func (actor *ValidatorSchema) GetChainsTableSchema() string                           { panic(ValPan) }
func (actor *ValidatorSchema) GetChainsQuery(_ string, _ int64) string                { panic(ValPan) }
func (actor *ValidatorSchema) ClearAllChainsQuery() string                            { panic(ValPan) }
