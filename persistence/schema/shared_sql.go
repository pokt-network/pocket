package schema

import (
	"bytes"
	"fmt"
)

const (
	// We use `-1` with semantic variable names to indicate non-existence or non-validity
	// in various contexts to avoid the usage of nullability in columns and for performance
	// optimization purposes.
	DefaultBigInt = -1

	// Common SQL selectors
	AllColsSelector  = "*"
	AnyValueSelector = "1"

	// Common column names
	AddressCol         = "address"
	BalanceCol         = "balance"
	PublicKeyCol       = "public_key"
	NameCol            = "name"
	StakedTokensCol    = "staked_tokens"
	ServiceURLCol      = "service_url"
	OutputAddressCol   = "output_address"
	UnstakingHeightCol = "unstaking_height"
	PausedHeightCol    = "paused_height"
	ChainIDCol         = "chain_id"
	MaxRelaysCol       = "max_relays"
	HeightCol          = "height"
)

func ProtocolActorTableSchema(actorSpecificColName, constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s BIGINT NOT NULL default %d,
			%s BIGINT NOT NULL default %d,
			%s BIGINT NOT NULL default %d,

			CONSTRAINT %s UNIQUE (%s, %s)
		)`,
		AddressCol,
		PublicKeyCol,
		StakedTokensCol,
		actorSpecificColName,
		OutputAddressCol,
		PausedHeightCol,
		DefaultBigInt,
		UnstakingHeightCol,
		DefaultBigInt,
		HeightCol,
		DefaultBigInt,
		constraintName,
		AddressCol,
		HeightCol)
}

func ProtocolActorChainsTableSchema(constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s CHAR(4) NOT NULL,
			%s BIGINT NOT NULL default %d,

			CONSTRAINT %s UNIQUE (%s, %s, %s)
		)`, AddressCol, ChainIDCol, HeightCol, DefaultBigInt, constraintName, AddressCol, ChainIDCol, HeightCol)
}

func Select(selector, address string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		selector, tableName, address, height)
}

// INCOMPLETE: Since we do not support `DeleteActor`, we do not filter on status or deletion here.
//             The onus of filtering is currently on the client but may be changed in the future.
func SelectActors(actorSpecificParam string, height int64, tableName string) string {
	return fmt.Sprintf(`
			SELECT DISTINCT ON (address) address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height
			FROM %s
			WHERE height<=%d
			ORDER BY address, height DESC
       `, actorSpecificParam, tableName, height)
}

func SelectChains(selector, address string, height int64, actorTableName, chainsTableName string) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE address='%s' AND height=(%s);`,
		selector, chainsTableName, address, Select(HeightCol, address, height, actorTableName))
}

func Exists(address string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT EXISTS(%s)`, Select(AnyValueSelector, address, height, tableName))
}

// Explainer:
//   (SELECT MAX(height), address FROM %s GROUP BY address) ->
//       returns latest/max height for each address
//   (height, address) IN (SELECT MAX(height), address FROM %s GROUP BY address) ->
//       ensures the query is acting on max height for the addresses
func ReadyToUnstake(unstakingHeight int64, tableName string) string {
	return fmt.Sprintf(`
		SELECT address, staked_tokens, output_address
		FROM %s WHERE unstaking_height=%d
			AND (height, address) IN (SELECT MAX(height), address FROM %s GROUP BY address)`,
		tableName, unstakingHeight, tableName)
}

func Insert(
	actor BaseActor,
	actorSpecificParam, actorSpecificParamValue,
	constraintName, chainsConstraintName,
	tableName, chainsTableName string,
	height int64) string {
	insertStatement := fmt.Sprintf(
		`INSERT INTO %s (address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
				VALUES('%s', '%s', '%s', '%s', '%s', %d, %d, %d)
				ON CONFLICT ON CONSTRAINT %s
				DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, %s=EXCLUDED.%s,
							  paused_height=EXCLUDED.paused_height, unstaking_height=EXCLUDED.unstaking_height,
							  height=EXCLUDED.height`,
		tableName, actorSpecificParam,
		actor.Address, actor.PublicKey, actor.StakedTokens, actorSpecificParamValue,
		actor.OutputAddress, actor.PausedHeight, actor.UnstakingHeight, height,
		constraintName,
		actorSpecificParam, actorSpecificParam)

	if actor.Chains == nil {
		return insertStatement
	}

	return fmt.Sprintf("WITH baseTableInsert AS (%s)\n%s",
		insertStatement, InsertChains(actor.Address, actor.Chains, height, chainsTableName, chainsConstraintName))
}

func InsertChains(address string, chains []string, height int64, tableName, constraintName string) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (address, chain_id, height) VALUES", tableName))

	maxIndex := len(chains) - 1
	for i, chain := range chains {
		buffer.WriteString(fmt.Sprintf("\n('%s', '%s', %d)", address, chain, height))
		if i < maxIndex {
			buffer.WriteString(",")
		}
	}

	buffer.WriteString(fmt.Sprintf("\nON CONFLICT ON CONSTRAINT %s DO NOTHING", constraintName))

	return buffer.String()
}

func Update(address, stakedTokens, actorSpecificParam, actorSpecificParamValue string, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(
		`INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, '%s', '%s', output_address, paused_height, unstaking_height, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		    ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, %s=EXCLUDED.%s, height=EXCLUDED.height`,
		tableName, actorSpecificParam,
		stakedTokens, actorSpecificParamValue, height,
		tableName, address, height,
		constraintName,
		actorSpecificParam, actorSpecificParam)
}

func UpdateUnstakingHeight(address, actorSpecificParam string, unstakingHeight, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
		(
			SELECT address, public_key, staked_tokens, %s, output_address, paused_height, %d, %d
			FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
		)
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET unstaking_height=EXCLUDED.unstaking_height, height=EXCLUDED.height`,
		tableName, actorSpecificParam,
		actorSpecificParam, unstakingHeight, height,
		tableName, address, height,
		constraintName)
}

func UpdateStakeAmount(address, actorSpecificParam, stakeAmount string, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
		(
			SELECT address, public_key, '%s', %s, output_address, paused_height, unstaking_height, %d
			FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
		)
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, height=EXCLUDED.height`,
		tableName, actorSpecificParam,
		stakeAmount, actorSpecificParam, height,
		tableName, address, height,
		constraintName)
}

func UpdatePausedHeight(address, actorSpecificParam string, pausedHeight, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
		(
			SELECT address, public_key, staked_tokens, %s, output_address, %d, unstaking_height, %d
			FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
		)
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET paused_height=EXCLUDED.paused_height, height=EXCLUDED.height`,
		tableName, actorSpecificParam, actorSpecificParam,
		pausedHeight, height,
		tableName, address, height, constraintName)
}

func UpdateUnstakedHeightIfPausedBefore(actorSpecificParam string, unstakingHeight, pausedBeforeHeight, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s (address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
		(
			SELECT address, public_key, staked_tokens, %s, output_address, paused_height, %d, %d
			FROM %s WHERE paused_height<%d
				AND (height,address) IN (SELECT MAX(height),address from %s GROUP BY address)
        )
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET unstaking_height=EXCLUDED.unstaking_height`,
		tableName, actorSpecificParam,
		actorSpecificParam, unstakingHeight, height,
		tableName, pausedBeforeHeight,
		tableName,
		constraintName)
}

func NullifyChains(address string, height int64, tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE address='%s' AND height=%d", tableName, address, height)
}

// Exposed for debugging purposes only
func ClearAll(tableName string) string {
	return fmt.Sprintf(`DELETE FROM %s`, tableName)
}
