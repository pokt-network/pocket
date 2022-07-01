package schema

var _ ProtocolActorSchema = &ValidatorSchema{}

type ValidatorSchema struct {
	BaseProtocolActorSchema
}

var ValidatorActor ProtocolActorSchema = &ValidatorSchema{
	BaseProtocolActorSchema: BaseProtocolActorSchema{
		tableName:       "validator",
		chainsTableName: "", // intentionally empty

		actorSpecificColName: ServiceURLCol,

		heightConstraintName:       "validator_node_height",
		chainsHeightConstraintName: "", // intentionally empty
	},
}

func (actor *ValidatorSchema) GetChainsTableSchema() string {
	return ""
}

func (actor *ValidatorSchema) GetChainsQuery(address string, height int64) string {
	return ""
}

func (actor *ValidatorSchema) InsertQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(BaseActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
	},
		actor.actorSpecificColName, maxRelays,
		actor.heightConstraintName, "",
		actor.tableName, "",
		height)
}

func (actor *ValidatorSchema) UpdateChainsQuery(address string, chains []string, height int64) string {
	return ""
}

func (actor *ValidatorSchema) ClearAllChainsQuery() string {
	return ""
}
