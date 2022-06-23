package schema

const (
	AppTableName            = "app"
	AppHeightConstraintName = "app_height"

	AppChainsTableName            = "app_chains"
	AppChainsHeightConstraintName = "app_chain_height"
)

var (
	AppTableSchema       = GenericActorTableSchema(MaxRelaysCol, AppHeightConstraintName)
	AppChainsTableSchema = ChainsTableSchema(AppChainsHeightConstraintName)
)

/* Read Queries */

// Returns a query to retrieve all of a single Application's attributes.
func AppQuery(address string, height int64) string {
	return Select(AllColsSelector, address, height, AppTableName)
}

// Returns a query for the existence of an Application given its address.
func AppExistsQuery(address string, height int64) string {
	return Exists(address, height, AppTableName)
}

// Returns a query to retrieve data associated with all the apps ready to unstake.
func AppsReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(unstakingHeight, AppTableName)
}

// Returns a query to retrieve the output address of an application given its operator address.
// DISCUSS(drewsky): Why/how we even need this. What is an output & operator for an app?
func AppOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, AppTableName)
}

// Returns a query to retrieve the height at which an application was paused.
func AppPausedHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, AppTableName)
}

// Returns a query to retrieve the height at which an application started unstaking.
// DISCUSS(team): if current_height == unstaking_height - is the actor unstaking or unstaked (i.e. did we process the block yet => yes if you're a replica and no if you're a proposer)?
func AppUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeightCol, address, height, AppTableName)
}

// Returns a query to retrieve all the data associated with the chains an application is staked for.
func AppChainsQuery(address string, height int64) string {
	return SelectChains(AllColsSelector, address, height, AppTableName, AppChainsTableName)
}

/* Create Queries */

// Returns a query to create a new application with all of the necessary data.
func InsertAppQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(GenericActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	},
		MaxRelaysCol, maxRelays,
		AppHeightConstraintName, AppChainsHeightConstraintName,
		AppTableName, AppChainsTableName,
		height)
}

/* Update Queries */

// Returns a query to update an application's stake and/or max relays.
func UpdateAppQuery(address, stakedTokens, maxRelays string, height int64) string {
	return Update(address, stakedTokens, MaxRelaysCol, maxRelays, height, AppTableName, AppHeightConstraintName)
}

// Returns a query to update the chains an application is staked for.
func UpdateAppChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, AppChainsTableName, AppChainsHeightConstraintName)
}

// Returns a query to update the height at which an application is unstaking.
func UpdateAppUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, MaxRelaysCol, unstakingHeight, height, AppTableName, AppHeightConstraintName)
}

// Returns a query to update the height at which an application is paused.
func UpdateAppPausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, MaxRelaysCol, pausedHeight, height, AppTableName, AppHeightConstraintName)
}

// Returns a query to start unstaking applications which have been paused.
func UpdateAppsPausedBefore(pauseBeforeHeight, unstakingHeight, height int64) string {
	return UpdatePausedBefore(MaxRelaysCol, unstakingHeight, pauseBeforeHeight, height, AppTableName, AppHeightConstraintName)
}

/* Delete Queries - used debugging only */

// Deletes all the applications.
func ClearAllAppsQuery() string {
	return ClearAll(AppTableName)
}

// Deletes all the data associated with the chains that applications are staked for.
func ClearAllAppChainsQuery() string {
	return ClearAll(AppChainsTableName)
}
