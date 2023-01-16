package types

import (
	"fmt"
)

var _ ProtocolAccountSchema = &BaseProtocolAccountSchema{}

type BaseProtocolAccountSchema struct {
	// SQL Tables
	tableName string

	// SQL Columns
	accountSpecificColName string

	// SQL Constraints
	heightConstraintName string
}

func (account *BaseProtocolAccountSchema) GetTableName() string {
	return account.tableName
}

func (account *BaseProtocolAccountSchema) GetAccountSpecificColName() string {
	return account.accountSpecificColName
}

func (account *BaseProtocolAccountSchema) GetTableSchema() string {
	return protocolAccountTableSchema(account.accountSpecificColName, account.heightConstraintName)
}

func (account *BaseProtocolAccountSchema) GetAccountAmountQuery(address string, height int64) string {
	return SelectBalance(account.accountSpecificColName, address, height, account.tableName)
}

func (account *BaseProtocolAccountSchema) GetAccountsUpdatedAtHeightQuery(height int64) string {
	return SelectAtHeight(fmt.Sprintf("%s,%s", account.accountSpecificColName, BalanceCol), height, account.tableName)
}

func (account *BaseProtocolAccountSchema) GetAllQuery(height int64) string {
	return SelectAccounts(height, account.accountSpecificColName, account.tableName)
}

func (account *BaseProtocolAccountSchema) InsertAccountQuery(address, amount string, height int64) string {
	return InsertAcc(account.accountSpecificColName, address, amount, height, account.tableName, account.heightConstraintName)
}

func (account *BaseProtocolAccountSchema) ClearAllAccounts() string {
	return fmt.Sprintf(`DELETE FROM %s`, account.tableName)
}
