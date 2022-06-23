package schema

import (
	"fmt"
)

const (
	// We use `-1` with semantic variable names to indicate non-existence or non-validity
	// in various contexts to avoid the usage of nullability in columns and for performance
	// optimization purposes.
	DefaultUnstakingHeight = -1
	DefaultEndHeight       = -1
	// Common SQL selectors
	AllSelector = "*"
	OneSelector = "1"
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

type GenericActor struct {
	Address         string
	PublicKey       string
	StakedTokens    string
	GenericParam    string
	OutputAddress   string
	PausedHeight    int64
	UnstakingHeight int64
	Chains          []string
}

func GenericActorTableSchema(actorSpecificColName, constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s BIGINT NOT NULL default -1,
			%s BIGINT NOT NULL default -1,
			%s BIGINT NOT NULL default -1,

			CONSTRAINT %s UNIQUE (%s, %s)
		)`,
		AddressCol,
		PublicKeyCol,
		StakedTokensCol,
		actorSpecificColName,
		OutputAddressCol,
		PausedHeightCol,
		UnstakingHeightCol,
		HeightCol,
		constraintName,
		AddressCol,
		HeightCol)
}

func ChainsTableSchema(constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s CHAR(4) NOT NULL,
			%s BIGINT NOT NULL default -1,

			CONSTRAINT %s UNIQUE (%s, %s, %s)
		)`, AddressCol, ChainIDCol, HeightCol, constraintName, AddressCol, ChainIDCol, HeightCol)
}

func AccountOrPoolSchema(mainColName, constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s BIGINT NOT NULL,
		    CONSTRAINT %s UNIQUE (%s, %s)
		)`, mainColName, BalanceCol, HeightCol, constraintName, mainColName, HeightCol)
}

func Select(selector string, address string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1`, selector, tableName, address, height)
}

func SelectChains(selector string, address string, height int64, baseTableName, chainsTableName string) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE address='%s' AND height=
(%s);`,
		selector, chainsTableName, address, Select(HeightCol, address, height, baseTableName))
}

func Exists(address string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT EXISTS(%s)`, Select(OneSelector, address, height, tableName))
}

func ReadyToUnstake(tableName string, unstakingHeight int64) string {
	return fmt.Sprintf(`SELECT address, staked_tokens, output_address FROM %s WHERE unstaking_height=%d AND (height,address) IN (
        select MAX(height),address from %s GROUP BY address
)`, tableName, unstakingHeight, tableName)
}

func Insert(
	actor GenericActor,
	genericParamName,
	genericParamValue,
	constraintName,
	chainsConstraintName,
	tableName,
	chainsTableName string,
	height int64) string {
	// base table
	insertStatement := fmt.Sprintf(
		`INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
				VALUES('%s','%s','%s','%s','%s',%d,%d,%d)
				ON CONFLICT ON CONSTRAINT %s
					DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, %s=EXCLUDED.%s, paused_height=EXCLUDED.paused_height, unstaking_height=EXCLUDED.unstaking_height, height=EXCLUDED.height`,
		tableName, genericParamName,
		actor.Address, actor.PublicKey, actor.StakedTokens, genericParamValue,
		actor.OutputAddress, actor.PausedHeight, actor.UnstakingHeight,
		height, constraintName, genericParamName, genericParamName)
	if actor.Chains == nil {
		return insertStatement
	}
	return fmt.Sprintf("WITH ins1 AS (%s)\n%s",
		insertStatement, InsertChains(actor.Address, actor.Chains, height, chainsTableName, chainsConstraintName))
}

func InsertChains(address string, chains []string, height int64, tableName, constraintName string) string {
	insert := fmt.Sprintf("INSERT INTO %s (address, chain_id, height) VALUES", tableName)
	maxIndex := len(chains) - 1
	for i, chain := range chains {
		insert += fmt.Sprintf("\n('%s', '%s', %d)", address, chain, height)
		if i < maxIndex {
			insert += ","
		}
	}
	insert += fmt.Sprintf(`
     ON CONFLICT ON CONSTRAINT %s
     DO NOTHING`, constraintName)
	return insert
}

func NullifyChains(address string, height int64, tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE address='%s' AND height=%d", tableName, address, height)
}

func InsertAcc(genericParamName, genericParamValue, amount string, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s (%s, balance, height)
			VALUES ('%s','%s',%d)
			ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height
		`, tableName, genericParamName, genericParamValue, amount, height, constraintName)
}

func SelectBalance(genericParamName, genericParamValue string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE %s='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		tableName, genericParamName, genericParamValue, height)
}

func Update(address, stakedTokens, genericParamName, genericParamValue string, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(
		`INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, '%s', '%s', output_address, paused_height, unstaking_height, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		    ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET staked_tokens=EXCLUDED.staked_tokens, %s=EXCLUDED.%s, height=EXCLUDED.height`,
		tableName, genericParamName,
		stakedTokens, genericParamValue, height,
		tableName, address, height, constraintName, genericParamName, genericParamName)
}

func UpdateUnstakingHeight(address, genericParamName string, unstakingHeight, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, staked_tokens, %s, output_address, paused_height, %d, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET unstaking_height=EXCLUDED.unstaking_height, height=EXCLUDED.height`,
		tableName, genericParamName, genericParamName,
		unstakingHeight, height,
		tableName, address, height, constraintName)

}

func UpdatePausedHeight(address, genericParamName string, pausedHeight, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, staked_tokens, %s, output_address, %d, unstaking_height, %d
				FROM %s WHERE address='%s' AND height<=%d ORDER BY height DESC LIMIT 1
			)
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET paused_height=EXCLUDED.paused_height, height=EXCLUDED.height`,
		tableName, genericParamName, genericParamName,
		pausedHeight, height,
		tableName, address, height, constraintName)
}

// DOCUMENT(team): Need to do a better job at documenting the process of paused apps being turned into unstaking apps.
//                 This pertains to all instances in the codebase, not just here.
func UpdatePausedBefore(genericParamName string, unstakingHeight, pausedBeforeHeight, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s(address, public_key, staked_tokens, %s, output_address, paused_height, unstaking_height, height)
			(
				SELECT address, public_key, staked_tokens, %s, output_address, paused_height, %d, %d
				FROM %s WHERE paused_height<%d AND paused_height>=-1 AND (height,address) IN (
        		  SELECT MAX(height),address from %s GROUP BY address
				)
		)
		ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET unstaking_height=EXCLUDED.unstaking_height`,
		tableName, genericParamName, genericParamName,
		unstakingHeight, height,
		tableName, pausedBeforeHeight, tableName, constraintName)
}

// Exposed for debugging purposes only
func ClearAll(tableName string) string {
	return fmt.Sprintf(`DELETE FROM %s`, tableName)
}
