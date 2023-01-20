package types

import (
	"fmt"
)

var _ ProtocolAccountSchema = &baseProtocolAccountSchema{}

// Implements the ProtocolAccountSchema interface that can be shared across both Accounts and Pools
// allowing for the generalisation and sharing of code between these two entities
type baseProtocolAccountSchema struct {
	// SQL Tables
	tableName string

	// SQL Columns
	accountSpecificColName string

	// SQL Constraints
	heightConstraintName string
}

func (account baseProtocolAccountSchema) GetTableName() string {
	return account.tableName
}

func (account baseProtocolAccountSchema) GetAccountSpecificColName() string {
	return account.accountSpecificColName
}

func (account baseProtocolAccountSchema) GetTableSchema() string {
	return protocolAccountTableSchema(account.accountSpecificColName, account.heightConstraintName)
}

func (account baseProtocolAccountSchema) GetAccountAmountQuery(identifier string, height int64) string {
	return SelectBalance(account.accountSpecificColName, identifier, height, account.tableName)
}

func (account baseProtocolAccountSchema) GetAccountsUpdatedAtHeightQuery(height int64) string {
	return SelectAtHeight(fmt.Sprintf("%s,%s", account.accountSpecificColName, BalanceCol), height, account.tableName)
}

func (account baseProtocolAccountSchema) GetAllQuery(height int64) string {
	return SelectAccounts(height, account.accountSpecificColName, account.tableName)
}

func (account baseProtocolAccountSchema) InsertAccountQuery(identifier, amount string, height int64) string {
	return InsertAccount(account.accountSpecificColName, identifier, amount, height, account.tableName, account.heightConstraintName)
}

func (account baseProtocolAccountSchema) ClearAllAccounts() string {
	return fmt.Sprintf(`DELETE FROM %s`, account.tableName)
}
