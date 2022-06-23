package schema

// TODO (Team) NOTE: omitting 'missed blocks' for fear of creating a new record every time a validator misses a block
// TODO - likely will use block store and *byzantine validators* to process

const (
	ValTableName            = "validator"
	ValidatorConstraintName = "val_height"
)

var (
	ValTableSchema = TableSchema(ServiceURLCol, ValidatorConstraintName)
)

func ValidatorQuery(address string, height int64) string {
	return Select(AllSelector, address, height, ValTableName)
}

func ValidatorOutputAddressQuery(operatorAddress string, height int64) string {
	return Select(OutputAddressCol, operatorAddress, height, ValTableName)
}

func ValidatorUnstakingHeightQuery(address string, height int64) string {
	return Select(UnstakingHeightCol, address, height, ValTableName)
}

func ValidatorPauseHeightQuery(address string, height int64) string {
	return Select(PausedHeightCol, address, height, ValTableName)
}

func ValidatorExistsQuery(address string, height int64) string {
	return Exists(address, height, ValTableName)
}

func ValidatorReadyToUnstakeQuery(unstakingHeight int64) string {
	return ReadyToUnstake(ValTableName, unstakingHeight)
}

func InsertValidatorQuery(address, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight int64, _ []string, height int64) string {
	return Insert(GenericActor{
		Address:         address,
		PublicKey:       publicKey,
		StakedTokens:    stakedTokens,
		OutputAddress:   outputAddress,
		PausedHeight:    pausedHeight,
		UnstakingHeight: unstakingHeight,
	}, ServiceURLCol, serviceURL, ValidatorConstraintName, "", ValTableName, "", height)
}

func UpdateValidatorQuery(address, stakedTokens, serviceURL string, height int64) string {
	return Update(address, stakedTokens, ServiceURLCol, serviceURL, height, ValTableName, ValidatorConstraintName)
}

func UpdateValidatorUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return UpdateUnstakingHeight(address, ServiceURLCol, unstakingHeight, height, ValTableName, ValidatorConstraintName)
}

func UpdateValidatorPausedHeightQuery(address string, pausedHeight, height int64) string {
	return UpdatePausedHeight(address, ServiceURLCol, pausedHeight, height, ValTableName, ValidatorConstraintName)
}

func UpdateValidatorsPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return UpdatePausedBefore(ServiceURLCol, unstakingHeight, pauseBeforeHeight, currentHeight, ValTableName, ValidatorConstraintName)
}

func ClearAllValidatorsQuery() string {
	return ClearAll(ValTableName)
}
