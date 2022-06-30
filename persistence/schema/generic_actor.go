package schema

type GenericActor struct {
	Address            string
	PublicKey          string
	StakedTokens       string
	ActorSpecificParam string // IMPROVE: May need to be refactored or converted to a list
	OutputAddress      string
	PausedHeight       int64
	UnstakingHeight    int64
	Chains             []string // IMPROVE: Consider creating a `type Chain string` for chains
}

var _ ProtocolActorSchema = &GenericProtocolActor{}

// Implements the ProtocolActor with behaviour that can be embedded (i.e. inherited) by other protocol
// actors for a share implementation.
//
// Note that this implementation assumes the protocol actor is chain dependant, so that behaviour needs
// to be overridden if the actor (e.g. Validator) is chain independent.
type GenericProtocolActor struct {
	// SQL Tables
	tableName       string
	chainsTableName string

	// SQL Columns
	actorSpecificColName string // CONSIDERATION: If actor specific behaviour expands, this will need to be refactored to be a list.

	// SQL Constraints
	heightConstraintName       string
	chainsHeightConstraintName string
}

func (actor *GenericProtocolActor) GetTableName() string {
	return actor.tableName
}

func (actor *GenericProtocolActor) GetChainsTableName() string {
	return actor.chainsTableName
}

func (actor *GenericProtocolActor) GetActorSpecificColName() string {
	return actor.actorSpecificColName
}

func (actor *GenericProtocolActor) GetTableSchema() string {
	return ProtocolActorTableSchema(actor.actorSpecificColName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) GetChainsTableSchema() string {
	return ProtocolActorChainsTableSchema(actor.chainsHeightConstraintName)
}

func (actor *GenericProtocolActor) GetQuery(address string, height int64) string {
	return Select(AllColsSelector, address, height, actor.tableName)
}

func (actor *GenericProtocolActor) GetExistsQuery(address string, height int64) string {
	return Exists(address, height, actor.tableName)
}

func (actor *GenericProtocolActor) GetReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(unstakingHeight, actor.tableName)
}

func (actor *GenericProtocolActor) GetOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, actor.tableName)
}

func (actor *GenericProtocolActor) GetPausedHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, actor.tableName)
}

func (actor *GenericProtocolActor) GetUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeightCol, address, height, actor.tableName)
}

func (actor *GenericProtocolActor) GetChainsQuery(address string, height int64) string {
	return SelectChains(AllColsSelector, address, height, actor.tableName, actor.chainsTableName)
}

func (actor *GenericProtocolActor) InsertQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(GenericActor{
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

func (actor *GenericProtocolActor) UpdateQuery(address, stakedTokens, maxRelays string, height int64) string {
	return Update(address, stakedTokens, actor.actorSpecificColName, maxRelays, height, actor.tableName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) UpdateChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, actor.chainsTableName, actor.chainsHeightConstraintName)
}

func (actor *GenericProtocolActor) UpdateUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, actor.actorSpecificColName, unstakingHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) UpdatePausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, actor.actorSpecificColName, pausedHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) UpdateUnstakedHeightIfPausedBeforeQuery(pauseBeforeHeight, unstakingHeight, height int64) string {
	return UpdateUnstakedHeightIfPausedBefore(actor.actorSpecificColName, unstakingHeight, pauseBeforeHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) ClearAllQuery() string {
	return ClearAll(actor.tableName)
}

func (actor *GenericProtocolActor) ClearAllChainsQuery() string {
	return ClearAll(actor.chainsTableName)
}
