package schema

import "fmt"

// TODO (Team) NOTE: omitting 'missed blocks' for fear of creating a new record every time a validator misses a block
// TODO - likely will use block store and *byzantine validators* to process

const (
	ValTableName   = "validator"
	ValTableSchema = `(
			address          TEXT NOT NULL, /*look into this being a "computed" field*/
			public_key       TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			service_url      TEXT NOT NULL,
			output_address   TEXT  NOT NULL,
			paused_height    BIGINT NOT NULL default -1,
			unstaking_height BIGINT NOT NULL default -1,
			end_height       BIGINT NOT NULL
		)`
)

func ValidatorQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=%d`, ValTableName, address, DefaultEndHeight)
}

func ValidatorExistsQuery(address string) string {
	return fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE address='%s' AND end_height=%d)`, ValTableName, address, DefaultUnstakingHeight)
}

func ValidatorReadyToUnstakeQuery(unstakingHeight int64) string {
	return fmt.Sprintf(`SELECT address,staked_tokens,output_address FROM %s WHERE unstaking_height=%d`, ValTableName, unstakingHeight)
}

func ValidatorOutputAddressQuery(operatorAddress string) string {
	return fmt.Sprintf(`SELECT output_address FROM %s WHERE address='%s' AND end_height=%d`,
		ValTableName, operatorAddress, DefaultEndHeight)
}

func ValidatorUnstakingHeightQuery(address string, height int64) string { // TODO (Team) if current_height == unstaking_height - is the actor unstaking or unstaked? IE did we process the block yet?
	return fmt.Sprintf(`SELECT unstaking_height FROM %s WHERE address='%s' AND end_height=%d`,
		ValTableName, address, DefaultEndHeight)
}

func ValidatorPauseHeightQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT paused_height FROM %s WHERE address='%s' AND end_height=%d`,
		ValTableName, address, DefaultEndHeight)
}

func InsertValidatorQuery(address, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight, height int64) string {
	// insert into main table
	insertIntoValidatorTable := fmt.Sprintf(
		`INSERT INTO %s(address, public_key, staked_tokens, service_url, output_address, paused_height, unstaking_height, end_height)
				VALUES('%s','%s','%s','%s','%s',%d,%d,%d)`,
		ValTableName, address, publicKey, stakedTokens, serviceURL, outputAddress, pausedHeight, unstakingHeight, DefaultEndHeight)
	return insertIntoValidatorTable
}

func NullifyValidatorQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		ValTableName, height, address, DefaultEndHeight)
}

func UpdateValidatorQuery(address, stakedTokens, serviceURL string, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,'%s','%s',output_address,paused_height,unstaking_height,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		ValTableName, stakedTokens, serviceURL, DefaultEndHeight, ValTableName, address, height)
}

func UpdateValidatorUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,service_url,output_address,paused_height,%d,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		ValTableName, unstakingHeight, DefaultEndHeight, ValTableName, address, height)
}

func UpdateValidatorPausedHeightQuery(address string, pausedHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,service_url,output_address,%d,unstaking_height,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		ValTableName, pausedHeight, DefaultEndHeight, ValTableName, address, height)
}

func UpdateValidatorsPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return fmt.Sprintf(`INSERT INTO %s
	(address, public_key, staked_tokens, service_url, output_address, paused_height, unstaking_height, end_height)
	SELECT address, public_key, staked_tokens, service_url, output_address, paused_height, %d, %d
	FROM %s WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`, ValTableName, unstakingHeight, DefaultEndHeight, ValTableName, pauseBeforeHeight, currentHeight)
}

func NullifyValidatorsPausedBeforeQuery(pausedBeforeHeight, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`,
		ValTableName, height, pausedBeforeHeight, DefaultEndHeight)
}

func ClearAllValidatorsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ValTableName)
}
