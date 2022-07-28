package schema

// Interface common to all protocol actors at the persistence schema layer. This exposes SQL specific
// attributes and queries.
type ProtocolActorSchema interface {
	/*** Protocol Actor Attributes ***/

	// SQL Table Names
	GetTableName() string
	GetChainsTableName() string
	// SQL Table Schemas
	GetTableSchema() string
	GetChainsTableSchema() string
	// SQL Column Names
	GetActorSpecificColName() string

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
	// Returns a query to retrieve the stake amount of an actor
	GetStakeAmountQuery(address string, height int64) string
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
	UpdateUnstakedHeightIfPausedBeforeQuery(pauseBeforeHeight, unstakingHeight, height int64) string
	// Returns a query to update the actor's stake amount
	SetStakeAmountQuery(address string, stakeAmount string, height int64) string

	/*** Delete Queries - used debugging only /***/

	// Deletes all the Actors.
	ClearAllQuery() string
	// Deletes all the data associated with the chains that Actors are staked for.
	ClearAllChainsQuery() string
}
