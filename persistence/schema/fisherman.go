package schema

import "fmt"

const (
	// TODO (Team) consider Fisherman paused_height for paused bool - only if we can use the 'height' field by not allowing edit stakes during pause
	// TODO (Team) can we make address field computed?
	FishTableName   = "fisherman"
	FishTableSchema = `(
			address          TEXT NOT NULL,
			public_key       TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			service_url      TEXT NOT NULL,
			output_address   TEXT NOT NULL,
			paused_height    BIGINT NOT NULL default -1,
			unstaking_height BIGINT NOT NULL default -1,
			end_height       BIGINT NOT NULL default -1
		)`
	FishChainsTableName   = "fisherman_chains"
	FishChainsTableSchema = `(
			address      TEXT NOT NULL,
			chain_id     CHAR(4) NOT NULL,
			end_height   BIGINT NOT NULL default -1
		)`
)

func FishermanQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=(%d)`, FishTableName, address, DefaultEndHeight)
}

func FishermanChainsQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=(%d)`, FishChainsTableName, address, DefaultEndHeight)
}

func FishermanExistsQuery(address string) string {
	return fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE address='%s')`, FishTableName, address)
}

func FishermanReadyToUnstakeQuery(unstakingHeight int64) string {
	return fmt.Sprintf(`SELECT address,staked_tokens,output_address FROM %s WHERE unstaking_height=%d`, FishTableName, unstakingHeight)
}

func FishermanOutputAddressQuery(operatorAddress string) string {
	return fmt.Sprintf(`SELECT output_address FROM %s WHERE address='%s' AND end_height=%d`,
		FishTableName, operatorAddress, DefaultEndHeight)
}

func FishermanUnstakingHeightQuery(address string, height int64) string { // TODO (Team) if current_height == unstaking_height - is the actor unstaking or unstaked? IE did we process the block yet?
	return fmt.Sprintf(`SELECT unstaking_height FROM %s WHERE address='%s' AND end_height=%d`,
		FishTableName, address, DefaultEndHeight)
}

func FishermanPauseHeightQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT paused_height FROM %s WHERE address='%s' AND end_height=%d`,
		FishTableName, address, DefaultEndHeight)
}

func InsertFishermanQuery(address, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight, height int64, chains []string) string {
	maxIndex := len(chains) - 1
	// insert into main table
	insertIntoFishermanTable := fmt.Sprintf(
		`WITH ins1 AS (INSERT INTO %s(address, public_key, staked_tokens, service_url, output_address, paused_height, unstaking_height, end_height)
				VALUES('%s','%s','%s','%s','%s',%d,%d,%d) RETURNING address)`,
		FishTableName, address, publicKey, stakedTokens, serviceURL, outputAddress, pausedHeight, unstakingHeight, DefaultEndHeight)
	// insert into chains table for each chain
	insertIntoFishermanTable += "\nINSERT INTO Service_node_chains (address, chain_id, end_height) VALUES"
	for i, chain := range chains {
		insertIntoFishermanTable += fmt.Sprintf("\n((SELECT address FROM ins1), '%s', %d)", chain, DefaultEndHeight)
		if i < maxIndex {
			insertIntoFishermanTable += ","
		}
	}
	return insertIntoFishermanTable
}

func NullifyFishermanQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		FishTableName, height, address, DefaultEndHeight)
}

func NullifyFishermanChainsQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		FishChainsTableName, height, address, DefaultEndHeight)
}

func UpdateFishermanQuery(address, stakedTokens, serviceURL string, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,'%s','%s',output_address,paused_height,unstaking_height,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		FishTableName, stakedTokens, serviceURL, DefaultEndHeight, FishTableName, address, height)
}

func UpdateFishermanUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,service_url,output_address,paused_height,%d,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		FishTableName, unstakingHeight, DefaultEndHeight, FishTableName, address, height)
}

func UpdateFishermanPausedHeightQuery(address string, pausedHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,service_url,output_address,%d,unstaking_height,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		FishTableName, pausedHeight, DefaultEndHeight, FishTableName, address, height)
}

func UpdateFishermansPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return fmt.Sprintf(`INSERT INTO %s
	(address, public_key, staked_tokens, service_url, output_address, paused_height, unstaking_height, end_height)
	SELECT address, public_key, staked_tokens, service_url, output_address, paused_height, %d, %d
	FROM %s WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`, FishTableName, unstakingHeight, DefaultEndHeight, FishTableName, pauseBeforeHeight, currentHeight)
}

func NullifyFishermansPausedBeforeQuery(pausedBeforeHeight, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`,
		FishTableName, height, pausedBeforeHeight, DefaultEndHeight)
}

func UpdateFishermanChainsQuery(address string, chains []string, height int64) string {
	insert := fmt.Sprintf("\nINSERT INTO %s (address, chain_id, end_height) VALUES", FishChainsTableName)
	maxIndex := len(chains) - 1
	for i, chain := range chains {
		insert += fmt.Sprintf("\n('%s', '%s', %d)", address, chain, DefaultEndHeight)
		if i < maxIndex {
			insert += ","
		}
	}
	return insert
}

func ClearAllFishermanQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, FishTableName)
}

func ClearAllFishermanChainsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, FishChainsTableName)
}
