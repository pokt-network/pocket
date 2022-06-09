package schema

import "fmt"

const (
	// We use `-1` with semantic variable names to indicate non-existence or non-validity
	// in various contexts to avoid the usage of nullability in columns and for performance
	// optimization purposes.
	DefaultUnstakingHeight = -1 // TODO(team): Move this into a shared file?
	DefaultEndHeight       = -1 // TODO(team): Move this into a shared file?

	// DISCUSS(drewsky): How do we handle historical queries here? E.g. get staked chains at some specific height?
	AppTableName   = "app"
	AppTableSchema = `(
			address    	     TEXT NOT NULL,
			public_key 		 TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			max_relays		 TEXT NOT NULL,
			output_address   TEXT NOT NULL,
			paused_height    BIGINT NOT NULL default -1,
			unstaking_height BIGINT NOT NULL default -1,
			height       BIGINT NOT NULL default -1,

			CONSTRAINT app_height UNIQUE (address, height)
		)`

	AppChainsTableName   = "app_chains"
	AppChainsTableSchema = `(
			address      TEXT NOT NULL,
			chain_id     CHAR(4) NOT NULL,
			height   BIGINT NOT NULL default -1,

			CONSTRAINT app_chain_height UNIQUE (address, chain_id, height)
		)`
)

func AppQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`, AppTableName, address, height)
}

func AppChainsQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`, AppChainsTableName, address, height)
}

func AppExistsQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE address='%s' AND staked_tokens!='0' AND height<=%d ORDER BY height DESC LIMIT 1)`, AppTableName, address, height)
}

// DISCUSS(drewsky): Do we not want to filter by `unstaking_height >= unstakingHeight here in case unstaking failed at the exact height?
func AppReadyToUnstakeQuery(unstakingHeight int64) string {
	return fmt.Sprintf(`SELECT address, staked_tokens, output_address FROM %s WHERE unstaking_height=%d`, AppTableName, unstakingHeight)
}

func AppOutputAddressQuery(operatorAddress string, height int64) string {
	return fmt.Sprintf(`SELECT output_address FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		AppTableName, operatorAddress, height)
}

// DISCUSS(team): if current_height == unstaking_height - is the actor unstaking or unstaked
// (i.e. did we process the block yet => yes if you're a replica and no if you're a proposer)?
func AppUnstakingHeightQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT unstaking_height FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		AppTableName, address, height)
}

func AppPauseHeightQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT paused_height FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		AppTableName, address, height)
}

func InsertAppQuery(address, publicKey, stakedTokens, maxRelays, outputAddress string, pausedHeight, unstakingHeight int64, chains []string, height int64) string {
	insertStatement := fmt.Sprintf(
		`WITH ins1 AS (INSERT INTO %s(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, height)
				VALUES('%s','%s','%s','%s','%s',%d,%d,%d)
				ON CONFLICT ON CONSTRAINT app_height
					DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, max_relays=EXCLUDED.max_relays, paused_height=EXCLUDED.paused_height, unstaking_height=EXCLUDED.unstaking_height, height=EXCLUDED.height)`,
		AppTableName, address, publicKey, stakedTokens, maxRelays, outputAddress, pausedHeight, unstakingHeight, height)
	return fmt.Sprintf("%s\n%s", insertStatement, InsertAppChainsQuery(address, chains, DefaultEndHeight))
}

func InsertAppChainsQuery(address string, chains []string, height int64) string {
	insert := fmt.Sprintf("INSERT INTO %s (address, chain_id, height) VALUES", AppChainsTableName)
	maxIndex := len(chains) - 1
	for i, chain := range chains {
		insert += fmt.Sprintf("\n('%s', '%s', %d)", address, chain, height)
		if i < maxIndex {
			insert += ","
		}
	}
	insert += `
   ON CONFLICT ON CONSTRAINT app_chain_height
     DO UPDATE SET chain_id=EXCLUDED.chain_id`
	return insert
}

func UpdateAppQuery(address, stakedTokens, maxRelays string, height int64) string {
	return fmt.Sprintf(
		`INSERT INTO %s(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, '%s', '%s', output_address, paused_height, unstaking_height, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		ON CONFLICT ON CONSTRAINT app_height
			DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, max_relays=EXCLUDED.max_relays, height=EXCLUDED.height`,
		AppTableName,
		stakedTokens, maxRelays, height,
		AppTableName, address, height)
}

func UpdateAppChainsQuery(address string, chains []string, height int64) string {
	return InsertAppChainsQuery(address, chains, height)
}

func UpdateAppUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, staked_tokens, max_relays, output_address, paused_height, %d, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		ON CONFLICT ON CONSTRAINT app_height
			DO UPDATE SET unstaking_height=EXCLUDED.unstaking_height, height=EXCLUDED.height`,
		AppTableName,
		unstakingHeight, height,
		AppTableName, address, height)

}

func UpdateAppPausedHeightQuery(address string, pausedHeight, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, staked_tokens, max_relays, output_address, %d, unstaking_height, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		ON CONFLICT ON CONSTRAINT app_height
			DO UPDATE SET paused_height=EXCLUDED.paused_height, height=EXCLUDED.height`,
		AppTableName,
		pausedHeight, height,
		AppTableName, address, height)
}

// DOCUMENT(team): Need to do a better job at documenting the process of paused apps being turned into unstaking apps.
//                 This pertains to all instances in the codebase, not just here.
func UpdateAppsPausedBefore(pauseBeforeHeight, unstakingHeight, height int64) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, max_relays, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, staked_tokens, max_relays, output_address, paused_height, %d, %d
				FROM %s WHERE paused_height<%d AND paused_height>=0
			)
		ON CONFLICT ON CONSTRAINT app_height
			DO UPDATE SET unstaking_height=EXCLUDED.unstaking_height`,
		AppTableName,
		unstakingHeight, height,
		AppTableName, pauseBeforeHeight)
}

// Exposed for debugging purposes only
func ClearAllAppsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, AppTableName)
}

// Exposed for debugging purposes only
func ClearAllAppChainsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, AppChainsTableName)
}
