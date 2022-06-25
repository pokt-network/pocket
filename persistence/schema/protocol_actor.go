package schema

type ProtocolActor interface {
	/*** Protocol Actor Attributes ***/

	// SQL Table Names
	GetTableName() string
	GetChainsTableName() string
	// SQL Table Schemas
	GetTableSchema() string
	GetChainsTableSchema() string

	/*** Read/Get Queries ***/

	// Returns a query to retrieve all of a single Actor's attributes.
	GetQuery(address string, height int64) string
	// Returns a query for the existence of an Actor given its address.
	GetExistsQuery(address string, height int64) string
	// Returns a query to retrieve data associated with all the apps ready to unstake.
	GetReadyToUnstakeQuery(unstakingHeight int64) string
	// Returns a query to retrieve the output address of an Actor given its operator address.
	// DISCUSS(drewsky): Why/how we even need this. What is an output & operator for an app?
	GetOutputAddressQuery(operatorAddress string, height int64) string
	// Returns a query to retrieve the height at which an Actor was paused.
	GetPausedHeightQuery(address string, height int64) string
	// Returns a query to retrieve the height at which an Actor started unstaking.
	// DISCUSS(team): if current_height == unstaking_height - is the Actor unstaking or unstaked (i.e. did we process the block yet => yes if you're a replica and no if you're a proposer)?
	GetUnstakingHeightQuery(address string, height int64) string
	// Returns a query to retrieve all the data associated with the chains an Actor is staked for.
	GetChainsQuery(address string, height int64) string

	/*** Create/Insert Queries ***/

	// Returns a query to create a new Actor with all of the necessary data.
	InsertQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string

	/*** Update Queries ***/
	// Returns a query to update an Actor's stake and/or max relays.
	UpdateQuery(address, stakedTokens, maxRelays string, height int64) string
	// Returns a query to update the chains an Actor is staked for.
	UpdateChainsQuery(address string, chains []string, height int64) string
	// Returns a query to update the height at which an Actor is unstaking.
	UpdateUnstakingHeightQuery(address string, unstakingHeight, height int64) string
	// Returns a query to update the height at which an Actor is paused.
	UpdatePausedHeightQuery(address string, pausedHeight, height int64) string
	// Returns a query to start unstaking Actors which have been paused.
	UpdatePausedBefore(pauseBeforeHeight, unstakingHeight, height int64) string

	/*** Delete Queries - used debugging only /***/

	// Deletes all the Actors.
	ClearAllQuery() string
	// Deletes all the data associated with the chains that Actors are staked for.
	ClearAllChainsQuery() string
}

var _ ProtocolActor = &GenericProtocolActor{}

type GenericProtocolActor struct {
	// SQL Tables
	tableName       string
	chainsTableName string

	// SQL Columns
	actorSpecificColName string // NOTE: If actor specific behaviour expands, this may need to be a list.

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

func (actor *GenericProtocolActor) GetTableSchema() string {
	return GenericActorTableSchema(actor.actorSpecificColName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) GetChainsTableSchema() string {
	return ChainsTableSchema(actor.chainsHeightConstraintName)
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

func (actor *GenericProtocolActor) UpdatePausedBefore(pauseBeforeHeight, unstakingHeight, height int64) string {
	return UpdatePausedBefore(actor.actorSpecificColName, unstakingHeight, pauseBeforeHeight, height, actor.tableName, actor.heightConstraintName)
}

func (actor *GenericProtocolActor) ClearAllQuery() string {
	return ClearAll(actor.tableName)
}

func (actor *GenericProtocolActor) ClearAllChainsQuery() string {
	return ClearAll(actor.chainsTableName)
}
