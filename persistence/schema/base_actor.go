package schema

// TECHDEBT: Consider moving this to a protobuf. This struct was created to make testing simple of protocol actors that
//           share most of the schema. We need to investigate if there's a better solution, or document this more appropriately and generalize
//           across the entire codebase.
//
// TODO (Team) -> convert to interface
// type BaseActor interface
// GetAddress() string
// SetAddress(string)
// GetPublicKey() string
// SetPublicKey(string)
// ...
// NOTE: requires modifying shared, so better to leave it alone until we reach some stability

type BaseActor struct {
	Address            string
	PublicKey          string
	StakedTokens       string
	ActorSpecificParam string // IMPROVE: May need to be refactored or converted to a list
	OutputAddress      string
	PausedHeight       int64
	UnstakingHeight    int64
	Chains             []string // IMPROVE: Consider creating a `type Chain string` for chains
}

var _ ProtocolActorSchema = &BaseProtocolActorSchema{}

// Implements the ProtocolActorSchema with behaviour that can be embedded (i.e. inherited) by other protocol
// actors for a share implementation.
//
// Note that this implementation assumes the protocol actor is chain dependant, so that behaviour needs
// to be overridden if the actor (e.g. Validator) is chain independent.
type BaseProtocolActorSchema struct {
	// SQL Tables
	tableName       string
	chainsTableName string

	// SQL Columns
	actorSpecificColName string // CONSIDERATION: If actor specific behaviour expands, this will need to be refactored to be a list.

	// SQL Constraints
	heightConstraintName       string
	chainsHeightConstraintName string
}

func (actor *BaseProtocolActorSchema) GetTableName() string {
	return actor.tableName
}

func (actor *BaseProtocolActorSchema) GetChainsTableName() string {
	return actor.chainsTableName
}

func (actor *BaseProtocolActorSchema) GetActorSpecificColName() string {
	return actor.actorSpecificColName
}

func (actor *BaseProtocolActorSchema) GetTableSchema() string {
	return ProtocolActorTableSchema(actor.actorSpecificColName, actor.heightConstraintName)
}

func (actor *BaseProtocolActorSchema) GetChainsTableSchema() string {
	return ProtocolActorChainsTableSchema(actor.chainsHeightConstraintName)
}

func (actor *BaseProtocolActorSchema) GetUpdatedAtHeightQuery(height int64) string {
	return SelectAtHeight(AllColsSelector, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetQuery(address string, height int64) string {
	return Select(AllColsSelector, address, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetExistsQuery(address string, height int64) string {
	return Exists(address, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(unstakingHeight, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetPausedHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeightCol, address, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetChainsQuery(address string, height int64) string {
	return SelectChains(AllColsSelector, address, height, actor.tableName, actor.chainsTableName)
}

func (actor *BaseProtocolActorSchema) InsertQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(BaseActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	},
		actor.actorSpecificColName, maxRelays,
		actor.heightConstraintName, actor.chainsHeightConstraintName,
		actor.tableName, actor.chainsTableName,
		height)
}

func (actor *BaseProtocolActorSchema) UpdateQuery(address, stakedTokens, maxRelays string, height int64) string {
	return Update(address, stakedTokens, actor.actorSpecificColName, maxRelays, height, actor.tableName, actor.heightConstraintName)
}

func (actor *BaseProtocolActorSchema) UpdateChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, actor.chainsTableName, actor.chainsHeightConstraintName)
}

func (actor *BaseProtocolActorSchema) UpdateUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, actor.actorSpecificColName, unstakingHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *BaseProtocolActorSchema) UpdatePausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, actor.actorSpecificColName, pausedHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *BaseProtocolActorSchema) UpdateUnstakedHeightIfPausedBeforeQuery(pauseBeforeHeight, unstakingHeight, height int64) string {
	return UpdateUnstakedHeightIfPausedBefore(actor.actorSpecificColName, unstakingHeight, pauseBeforeHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *BaseProtocolActorSchema) ClearAllQuery() string {
	return ClearAll(actor.tableName)
}

func (actor *BaseProtocolActorSchema) ClearAllChainsQuery() string {
	return ClearAll(actor.chainsTableName)
}
