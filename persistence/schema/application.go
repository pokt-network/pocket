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

func AppQuery(address string, height int64) string {
	return Select(AllSelector, address, height, AppTableName)
}

func AppExistsQuery(address string, height int64) string {
	return Exists(address, height, AppTableName)
}

func AppReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(AppTableName, unstakingHeight)
}

func AppOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, AppTableName)
}

func AppPauseHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, AppTableName)
}

// DISCUSS(team): if current_height == unstaking_height - is the actor unstaking or unstaked
// (i.e. did we process the block yet => yes if you're a replica and no if you're a proposer)?
func AppUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeightCol, address, height, AppTableName)
}

func AppChainsQuery(address string, height int64) string {
	return SelectChains(AllSelector, address, height, AppTableName, AppChainsTableName)
}

func InsertAppQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(GenericActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	}, MaxRelaysCol, maxRelays, AppHeightConstraintName, AppChainsHeightConstraintName, AppTableName, AppChainsTableName, height)
}

func UpdateAppQuery(address, stakedTokens, maxRelays string, height int64) string {
	return Update(address, stakedTokens, MaxRelaysCol, maxRelays, height, AppTableName, AppHeightConstraintName)
}

func UpdateAppChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, AppChainsTableName, AppChainsHeightConstraintName)
}

func UpdateAppUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, MaxRelaysCol, unstakingHeight, height, AppTableName, AppHeightConstraintName)

}

func UpdateAppPausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, MaxRelaysCol, pausedHeight, height, AppTableName, AppHeightConstraintName)
}

func UpdateAppsPausedBefore(pauseBeforeHeight, unstakingHeight, height int64) string {
	return UpdatePausedBefore(MaxRelaysCol, unstakingHeight, pauseBeforeHeight, height, AppTableName, AppHeightConstraintName)
}

func ClearAllAppsQuery() string {
	return ClearAll(AppTableName)
}

func ClearAllAppChainsQuery() string {
	return ClearAll(AppChainsTableName)
}
