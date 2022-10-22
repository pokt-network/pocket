package types

// IMPROVE: Move schema related functions to a separate sub-package
import "github.com/pokt-network/pocket/shared/modules"

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
	return SelectAtHeight(AddressCol, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetQuery(address string, height int64) string {
	return Select(AllColsSelector, address, height, actor.tableName)
}

func (actor *BaseProtocolActorSchema) GetAllQuery(height int64) string {
	return SelectActors(actor.GetActorSpecificColName(), height, actor.tableName)
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

func (actor *BaseProtocolActorSchema) GetStakeAmountQuery(address string, height int64) string {
	return Select(StakedTokensCol, address, height, actor.tableName)
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

func (actor *BaseProtocolActorSchema) InsertQuery(address, publicKey, stakedTokens, generic, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(&Actor{
		Address:         address,
		PublicKey:       publicKey,
		StakedAmount:    stakedTokens,
		Output:          outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	},
		actor.actorSpecificColName, generic,
		actor.heightConstraintName, actor.chainsHeightConstraintName,
		actor.tableName, actor.chainsTableName,
		height)
}

func (actor *BaseProtocolActorSchema) UpdateQuery(address, stakedTokens, generic string, height int64) string {
	return Update(address, stakedTokens, actor.actorSpecificColName, generic, height, actor.tableName, actor.heightConstraintName)
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

func (actor *BaseProtocolActorSchema) SetStakeAmountQuery(address string, stakedTokens string, height int64) string {
	return UpdateStakeAmount(address, actor.actorSpecificColName, stakedTokens, height, actor.tableName, actor.heightConstraintName)
}

func (actor *BaseProtocolActorSchema) ClearAllQuery() string {
	return ClearAll(actor.tableName)
}

func (actor *BaseProtocolActorSchema) ClearAllChainsQuery() string {
	return ClearAll(actor.chainsTableName)
}

var _ modules.Actor = &Actor{}

func (x *Actor) GetActorTyp() modules.ActorType {
	if x != nil {
		return x.GetActorType()
	}
	return ActorType_Undefined
}
