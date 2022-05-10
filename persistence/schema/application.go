package schema

import "fmt"

const (
	DefaultUnstakingHeight = -1
	DefaultEndHeight       = -1

	// TODO (team) look into address being a "computed" field
	AppTableName   = "app"
	AppTableSchema = `(
			address    	     TEXT NOT NULL,
			public_key 		 TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			max_relays		 TEXT NOT NULL,
			output_address   TEXT NOT NULL,
			paused_height    BIGINT NOT NULL default -1,
			unstaking_height BIGINT NOT NULL default -1,
			end_height       BIGINT NOT NULL default -1
		)`

	AppChainsTableName   = "app_chains"
	AppChainsTableSchema = `(
			address      TEXT NOT NULL,
			chain_id     CHAR(4) NOT NULL,
			end_height   BIGINT NOT NULL default -1
		)`
)

func AppQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=%d`, AppTableName, address, DefaultEndHeight)
}

func AppChainsQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=(%d)`, AppChainsTableName, address, DefaultEndHeight)
}

func AppExistsQuery(address string) string {
	return fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE address='%s')`, AppTableName, address)
}

func AppReadyToUnstakeQuery(unstakingHeight int64) string {
	return fmt.Sprintf(`SELECT address,staked_tokens,output_address FROM %s WHERE unstaking_height=%d`, AppTableName, unstakingHeight)
}

func AppOutputAddressQuery(operatorAddress string, height int64) string {
	return fmt.Sprintf(`SELECT output_address FROM %s WHERE address='%s' AND end_height=%d`,
		AppTableName, operatorAddress, DefaultEndHeight)
}

func AppUnstakingHeightQuery(address string, height int64) string { // TODO (Team) if current_height == unstaking_height - is the actor unstaking or unstaked? IE did we process the block yet?
	return fmt.Sprintf(`SELECT unstaking_height FROM %s WHERE address='%s' AND end_height=%d`,
		AppTableName, address, DefaultEndHeight)
}

func AppPauseHeightQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT paused_height FROM %s WHERE address='%s' AND end_height=%d`,
		AppTableName, address, DefaultEndHeight)
}

func InsertAppQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string) string {
	maxIndex := len(chains) - 1
	// insert into main table
	insertIntoAppTable := fmt.Sprintf(
		`WITH ins1 AS (INSERT INTO %s(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, end_height)
				VALUES('%s','%s','%s','%s','%s',%d,%d,%d) RETURNING address)`,
		AppTableName, address, publicKey, stakedTokens, maxRelays, outputAddress, pausedHeight, unstakingHeight, DefaultEndHeight)
	// insert into chains table for each chain
	insertIntoAppTable += "\nINSERT INTO app_chains (address, chain_id, end_height) VALUES"
	for i, chain := range chains {
		insertIntoAppTable += fmt.Sprintf("\n((SELECT address FROM ins1), '%s', %d)", chain, DefaultEndHeight)
		if i < maxIndex {
			insertIntoAppTable += ","
		}
	}
	return insertIntoAppTable
}

// https://www.postgresql.org/docs/current/sql-insert.html

func NullifyAppQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		AppTableName, height, address, DefaultEndHeight)
}

func NullifyAppChainsQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		AppChainsTableName, height, address, DefaultEndHeight)
}

func UpdateAppQuery(address, stakedTokens, maxRelays string, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,max_relays,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,'%s','%s',output_address,paused_height,unstaking_height,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		AppTableName, stakedTokens, maxRelays, DefaultEndHeight, AppTableName, address, height)
}

// func UpdateAppQuery(address, stakedTokens, maxRelays string, height int64) string {
// 	return fmt.Sprintf(`UPSERT INTO %s(address,public_key,staked_tokens,max_relays,output_address,paused_height,unstaking_height,end_height)
//                                ((SELECT address,public_key,'%s','%s',output_address,paused_height,unstaking_height,%d FROM %s WHERE address='%s'AND
//                                (end_height=%d OR end_height=(-1))))`,
// 		AppTableName, stakedTokens, maxRelays, DefaultEndHeight, AppTableName, address, height)
// }

func UpdateAppUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,max_relays,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,max_relays,output_address,paused_height,%d,%d FROM %s WHERE address='%s'AND
                               end_height=%d))`,
		AppTableName, unstakingHeight, DefaultEndHeight, AppTableName, address, height)
}

func UpdateAppPausedHeightQuery(address string, pauseHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,max_relays,output_address,paused_height,unstaking_height,end_height)
                               (SELECT address,public_key,staked_tokens,max_relays,output_address,%d,unstaking_height,%d FROM %s WHERE address='%s'AND
                               end_height=%d)`,
		AppTableName, pauseHeight, DefaultEndHeight, AppTableName, address, height)
}

func UpdateAppsPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return fmt.Sprintf(`INSERT INTO %s
	(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, end_height)
	SELECT address, public_key, staked_tokens, max_relays, output_address, paused_height, %d, %d
	FROM %s WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`, AppTableName, unstakingHeight, DefaultEndHeight, AppTableName, pauseBeforeHeight, currentHeight)
}

func NullifyAppsPausedBeforeQuery(pausedBeforeHeight, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`,
		AppTableName, height, pausedBeforeHeight, DefaultEndHeight)
}

func UpdateAppChainsQuery(address string, chains []string, height int64) string {
	insert := fmt.Sprintf("\nINSERT INTO %s (address, chain_id, end_height) VALUES", AppChainsTableName)
	maxIndex := len(chains) - 1
	for i, chain := range chains {
		insert += fmt.Sprintf("\n('%s', '%s', %d)", address, chain, DefaultEndHeight)
		if i < maxIndex {
			insert += ","
		}
	}
	return insert
}

func ClearAllAppQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, AppTableName)
}

func ClearAllAppChainsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, AppChainsTableName)
}
