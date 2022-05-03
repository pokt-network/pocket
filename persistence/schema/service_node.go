package schema

import "fmt"

const (
	ServiceNodeTableName   = "service_node"
	ServiceNodeTableSchema = `(
			address          TEXT NOT NULL, /*look into this being a "computed" field*/
			public_key       TEXT NOT NULL,
			staked_tokens    TEXT NOT NULL,
			service_url      TEXT NOT NULL,
			output_address   TEXT NOT NULL,
			paused_height    BIGINT NOT NULL default -1,
			unstaking_height BIGINT NOT NULL default -1,
			end_height       BIGINT NOT NULL default -1
		)`
	ServiceNodeChainsTableName   = "service_node_chains"
	ServiceNodeChainsTableSchema = `(
			address    TEXT NOT NULL,
			chain_id   CHAR(4),
			end_height BIGINT NOT NULL default -1
		)`
)

func ServiceNodeQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=%d`, ServiceNodeTableName, address, DefaultEndHeight)
}

func ServiceNodeChainsQuery(address string) string {
	return fmt.Sprintf(`SELECT * FROM %s WHERE address='%s' AND end_height=(%d)`, ServiceNodeChainsTableName, address, DefaultEndHeight)
}

func ServiceNodeExistsQuery(address string) string {
	return fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE address='%s')`, ServiceNodeTableName, address)
}

func ServiceNodeReadyToUnstakeQuery(unstakingHeight int64) string {
	return fmt.Sprintf(`SELECT address,staked_tokens,output_address FROM %s WHERE unstaking_height=%d`, ServiceNodeTableName, unstakingHeight)
}

func ServiceNodeOutputAddressQuery(operatorAddress string) string {
	return fmt.Sprintf(`SELECT output_address FROM %s WHERE address='%s' AND end_height=%d`,
		ServiceNodeTableName, operatorAddress, DefaultEndHeight)
}

func ServiceNodeUnstakingHeightQuery(address string, height int64) string { // TODO (Team) if current_height == unstaking_height - is the actor unstaking or unstaked? IE did we process the block yet?
	return fmt.Sprintf(`SELECT unstaking_height FROM %s WHERE address='%s' AND end_height=%d`,
		ServiceNodeTableName, address, DefaultEndHeight)
}

func ServiceNodePauseHeightQuery(address string, height int64) string {
	return fmt.Sprintf(`SELECT paused_height FROM %s WHERE address='%s' AND end_height=%d`,
		ServiceNodeTableName, address, DefaultEndHeight)
}

func InsertServiceNodeQuery(address, publicKey, stakedTokens, serviceURL, outputAddress string, pausedHeight, unstakingHeight, height int64, chains []string) string {
	maxIndex := len(chains) - 1
	// insert into main table
	insertIntoServiceNodeTable := fmt.Sprintf(
		`WITH ins1 AS (INSERT INTO %s(address, public_key, staked_tokens, service_url, output_address, paused_height, unstaking_height, end_height)
				VALUES('%s','%s','%s','%s','%s',%d,%d,%d) RETURNING address)`,
		ServiceNodeTableName, address, publicKey, stakedTokens, serviceURL, outputAddress, pausedHeight, unstakingHeight, DefaultEndHeight)
	// insert into chains table for each chain
	insertIntoServiceNodeTable += "\nINSERT INTO Service_node_chains (address, chain_id, end_height) VALUES"
	for i, chain := range chains {
		insertIntoServiceNodeTable += fmt.Sprintf("\n((SELECT address FROM ins1), '%s', %d)", chain, DefaultEndHeight)
		if i < maxIndex {
			insertIntoServiceNodeTable += ","
		}
	}
	return insertIntoServiceNodeTable
}

func NullifyServiceNodeQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		ServiceNodeTableName, height, address, DefaultEndHeight)
}

func NullifyServiceNodeChainsQuery(address string, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE address='%s' AND end_height=%d`,
		ServiceNodeChainsTableName, height, address, DefaultEndHeight)
}

func UpdateServiceNodeQuery(address, stakedTokens, serviceURL string, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,'%s','%s',output_address,paused_height,unstaking_height,%d FROM %s WHERE address='%s'AND 
                               end_height=%d))`,
		ServiceNodeTableName, stakedTokens, serviceURL, DefaultEndHeight, ServiceNodeTableName, address, height)
}

func UpdateServiceNodeUnstakingHeightQuery(address string, unstakingHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,service_url,output_address,paused_height,%d,%d FROM %s WHERE address='%s'AND 
                               end_height=%d))`,
		ServiceNodeTableName, unstakingHeight, DefaultEndHeight, ServiceNodeTableName, address, height)
}

func UpdateServiceNodePausedHeightQuery(address string, pauseHeight, height int64) string {
	return fmt.Sprintf(`INSERT INTO %s(address,public_key,staked_tokens,service_url,output_address,paused_height,unstaking_height,end_height)
                               ((SELECT address,public_key,staked_tokens,service_url,output_address,%d,unstaking_height,%d FROM %s WHERE address='%s'AND 
                               end_height=%d))`,
		ServiceNodeTableName, pauseHeight, DefaultEndHeight, ServiceNodeTableName, address, height)
}

func UpdateServiceNodesPausedBefore(pauseBeforeHeight, unstakingHeight, currentHeight int64) string {
	return fmt.Sprintf(`INSERT INTO %s
	(address, public_key, staked_tokens, service_url, output_address, paused_height, unstaking_height, end_height)
	SELECT address, public_key, staked_tokens, service_url, output_address, paused_height, %d, %d
	FROM %s WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`, ServiceNodeTableName, unstakingHeight, DefaultEndHeight, ServiceNodeTableName, pauseBeforeHeight, currentHeight)
}

func NullifyServiceNodesPausedBeforeQuery(pausedBeforeHeight, height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE paused_height<%d AND paused_height!=(-1) AND end_height=%d`,
		ServiceNodeTableName, height, pausedBeforeHeight, DefaultEndHeight)
}

func UpdateServiceNodeChainsQuery(address string, chains []string, height int64) string {
	insert := fmt.Sprintf("\nINSERT INTO %s (address, chain_id, end_height) VALUES", ServiceNodeChainsTableName)
	maxIndex := len(chains) - 1
	for i, chain := range chains {
		insert += fmt.Sprintf("\n('%s', '%s', %d)", address, chain, DefaultEndHeight)
		if i < maxIndex {
			insert += ","
		}
	}
	return insert
}

func ServiceNodeCountQuery(chain string, height int64) string {
	return fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE chain_id='%s' AND address IN (SELECT address FROM %s WHERE paused_height=(-1) AND end_height=(-1))`,
		ServiceNodeChainsTableName, chain, ServiceNodeTableName)
}

func ClearAllServiceNodesQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ServiceNodeTableName)
}

func ClearAllServiceNodesChainsQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ServiceNodeChainsTableName)
}
