package types

import "fmt"

const (
	AccountTableName        = "account"
	AccountHeightConstraint = "account_create_height"
	/*
		From the utility specification:
			A ModulePool is a particular type that though similar in structure to an Account, the
			functionality of each is quite specialized to its use case. These pools are maintained by
			the protocol and are completely autonomous, owned by no actor on the network. Unlike Accounts,
			tokens are able to be directly minted to and burned from ModulePools. Examples of ModuleAccounts
			include StakingPools and the FeeCollector
	*/
	PoolTableName        = "pool"
	PoolHeightConstraint = "pool_create_height"
)

var (
	AccountTableSchema = AccountOrPoolSchema(AddressCol, AccountHeightConstraint)
	PoolTableSchema    = AccountOrPoolSchema(NameCol, PoolHeightConstraint)
)

func GetAccountAmountQuery(address string, height int64) string {
	return SelectBalance(AddressCol, address, height, AccountTableName)
}

func InsertAccountAmountQuery(address, amount string, height int64) string {
	return InsertAcc(AddressCol, address, amount, height, AccountTableName, AccountHeightConstraint)
}

func GetPoolAmountQuery(name string, height int64) string {
	return SelectBalance(NameCol, name, height, PoolTableName)
}

func InsertPoolAmountQuery(name, amount string, height int64) string {
	return InsertAcc(NameCol, name, amount, height, PoolTableName, PoolHeightConstraint)
}

func AccountOrPoolSchema(mainColName, constraintName string) string {
	return fmt.Sprintf(`(
			%s TEXT NOT NULL,
			%s TEXT NOT NULL,
			%s BIGINT NOT NULL,

		    CONSTRAINT %s UNIQUE (%s, %s)
		)`, mainColName, BalanceCol, HeightCol, constraintName, mainColName, HeightCol)
}

func InsertAcc(actorSpecificParam, actorSpecificParamValue, amount string, height int64, tableName, constraintName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s (%s, balance, height)
			VALUES ('%s','%s',%d)
			ON CONFLICT ON CONSTRAINT %s
			DO UPDATE SET balance=EXCLUDED.balance, height=EXCLUDED.height
		`, tableName, actorSpecificParam, actorSpecificParamValue, amount, height, constraintName)
}

func SelectBalance(actorSpecificParam, actorSpecificParamValue string, height int64, tableName string) string {
	return fmt.Sprintf(`SELECT balance FROM %s WHERE %s='%s' AND height<=%d ORDER BY height DESC LIMIT 1`,
		tableName, actorSpecificParam, actorSpecificParamValue, height)
}

func SelectAccounts(height int64, tableName string) string {
	return fmt.Sprintf(`
			SELECT DISTINCT ON (address) address, balance, height
			FROM %s
			WHERE height<=%d
			ORDER BY address, height DESC
       `, tableName, height)
}

func SelectPools(height int64, tableName string) string {
	return fmt.Sprintf(`
			SELECT DISTINCT ON (name) name, balance, height
			FROM %s
			WHERE height<=%d
			ORDER BY name, height DESC
       `, tableName, height)
}

func ClearAllAccounts() string {
	return fmt.Sprintf(`DELETE FROM %s`, AccountTableName)
}

func ClearAllPools() string {
	return fmt.Sprintf(`DELETE FROM %s`, PoolTableName)
}
