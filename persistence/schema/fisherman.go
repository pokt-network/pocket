package schema

// TODO (Team) consider Fisherman paused_height for paused bool - only if we can use the 'height' field by not allowing edit stakes during pause
// TODO (Team) can we make Address field computed?

const (
	FishTableName            = "fisherman"
	FishChainsTableName      = "fisherman_chains"
	FishConstraintName       = "fish_height"
	FishChainsConstraintName = "fish_chain_height"
)

var (
	FishTableSchema       = GenericActorTableSchema(ServiceURLCol, FishConstraintName)
	FishChainsTableSchema = ChainsTableSchema(FishChainsConstraintName)
)

func FishermanQuery(address string, height int64) string {
	return Select(AllSelector, address, height, FishTableName)
}

func FishermanChainsQuery(address string, height int64) string {
	return SelectChains(AllSelector, address, height, FishTableName, FishChainsTableName)
}

func FishermanExistsQuery(address string, height int64) string {
	return Exists(address, height, FishTableName)
}

func FishermanReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(FishTableName, unstakingHeight)
}

func FishermanOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, FishTableName)
}

func FishermanUnstakingHeightQuery(address string, height int64) string { // TODO (Team) if current_height == unstaking_height - is the actor unstaking or unstaked? IE did we process the block yet?
	return Select(UnstakingHeightCol, address, height, FishTableName)
}

func FishermanPauseHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, FishTableName)
}

func InsertFishermanQuery(address, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(GenericActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	}, ServiceURLCol, serviceURL, FishConstraintName, FishChainsConstraintName, FishTableName, FishChainsTableName, height)
}

func UpdateFishermanQuery(address, stakedTokens, serviceURL string, height int64) string {
	return Update(address, stakedTokens, ServiceURLCol, serviceURL, height, FishTableName, FishConstraintName)
}

func UpdateFishermanUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, ServiceURLCol, unstakingHeight, height, FishTableName, FishConstraintName)
}

func UpdateFishermanPausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, ServiceURLCol, pausedHeight, height, FishTableName, FishConstraintName)
}

func UpdateFishermenPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return UpdatePausedBefore(ServiceURLCol, unstakingHeight, pauseBeforeHeight, currentHeight, FishTableName, FishConstraintName)
}

func UpdateFishermanChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, FishChainsTableName, FishChainsConstraintName)
}

func ClearAllFishermanQuery() string {
	return ClearAll(FishTableName)
}

func ClearAllFishermanChainsQuery() string {
	return ClearAll(FishChainsTableName)
}
