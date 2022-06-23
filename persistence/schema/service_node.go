package schema

const (
	ServiceNodeTableName            = "service_node"
	ServiceNodeChainsTableName      = "service_node_chains"
	ServiceNodeConstraintName       = "service_node_height"
	ServiceNodeChainsConstraintName = "service_node_chain_height"
)

var (
	ServiceNodeTableSchema       = GenericActorTableSchema(ServiceURLCol, ServiceNodeConstraintName)
	ServiceNodeChainsTableSchema = ChainsTableSchema(ServiceNodeChainsConstraintName)
)

func ServiceNodeQuery(address string, height int64) string {
	return Select(AllColsSelector, address, height, ServiceNodeTableName)
}

func ServiceNodeChainsQuery(address string, height int64) string {
	return SelectChains(AllColsSelector, address, height, ServiceNodeTableName, ServiceNodeChainsTableName)
}

func ServiceNodeExistsQuery(address string, height int64) string {
	return Exists(address, height, ServiceNodeTableName)
}

func ServiceNodeReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(unstakingHeight, ServiceNodeTableName)
}

func ServiceNodeOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, ServiceNodeTableName)
}

func ServiceNodeUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeightCol, address, height, ServiceNodeTableName)
}

func ServiceNodePauseHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, ServiceNodeTableName)
}

func InsertServiceNodeQuery(address, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	return Insert(GenericActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
		Chains:          chains,
	}, ServiceURLCol, serviceURL, ServiceNodeConstraintName, ServiceNodeChainsConstraintName, ServiceNodeTableName, ServiceNodeChainsTableName, height)
}

func UpdateServiceNodeQuery(address, stakedTokens, serviceURL string, height int64) string {
	return Update(address, stakedTokens, ServiceURLCol, serviceURL, height, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodeUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, ServiceURLCol, unstakingHeight, height, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodePausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, ServiceURLCol, pausedHeight, height, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodesPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return UpdatePausedBefore(ServiceURLCol, unstakingHeight, pauseBeforeHeight, currentHeight, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodeChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, ServiceNodeChainsTableName, ServiceNodeChainsConstraintName)
}

func ClearAllServiceNodesQuery() string {
	return ClearAll(ServiceNodeTableName)
}

func ClearAllServiceNodesChainsQuery() string {
	return ClearAll(ServiceNodeChainsTableName)
}
