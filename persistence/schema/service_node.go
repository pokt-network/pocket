package schema

const (
	ServiceNodeTableName            = "service_node"
	ServiceNodeChainsTableName      = "service_node_chains"
	ServiceNodeConstraintName       = "service_node_height"
	ServiceNodeChainsConstraintName = "service_node_chain_height"
)

var (
	ServiceNodeTableSchema       = TableSchema(ServiceURL, ServiceNodeConstraintName)
	ServiceNodeChainsTableSchema = ChainsTableSchema(ServiceNodeChainsConstraintName)
)

func ServiceNodeQuery(address string, height int64) string {
	return Select(AllSelector, address, height, ServiceNodeTableName)
}

func ServiceNodeChainsQuery(address string, height int64) string {
	return Select(AllSelector, address, height, ServiceNodeChainsTableName)
}

func ServiceNodeExistsQuery(address string, height int64) string {
	return Exists(address, height, ServiceNodeTableName)
}

func ServiceNodeReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(ServiceNodeTableName, unstakingHeight)
}

func ServiceNodeOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddress, operatorAddress, height, ServiceNodeTableName)
}

func ServiceNodeUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeight, address, height, ServiceNodeTableName)
}

func ServiceNodePauseHeightQuery(address string, height int64) string {
	return Select(PausedHeight, address, height, ServiceNodeTableName)
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
	}, ServiceURL, serviceURL, ServiceNodeConstraintName, ServiceNodeChainsConstraintName, ServiceNodeTableName, ServiceNodeChainsTableName, height)
}

func UpdateServiceNodeQuery(address, stakedTokens, serviceURL string, height int64) string {
	return Update(address, stakedTokens, ServiceURL, serviceURL, height, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodeUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, ServiceURL, unstakingHeight, height, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodePausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, ServiceURL, pausedHeight, height, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodesPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return UpdatePausedBefore(ServiceURL, unstakingHeight, pauseBeforeHeight, currentHeight, ServiceNodeTableName, ServiceNodeConstraintName)
}

func UpdateServiceNodeChainsQuery(address string, chains []string, height int64) string {
	return InsertChains(address, chains, height, ServiceNodeChainsTableName, ServiceNodeChainsConstraintName)
}

func ClearAllServiceNodesQuery() string {
	return ClearAll(ServiceNodeTableName)
}

func ClearAllServiceNodesChainsQuery() string {
	return ClearAll(ServiceNodeTableSchema)
}
