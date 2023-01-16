package types

// Interface common to protocol Accounts and Pools at the persistence schema layer.
// This exposes SQL specific attributes and queries.
type ProtocolAccountSchema interface {
	/*** Account Attributes ***/

	// SQL Table Name
	GetTableName() string
	// SQL Column Name
	GetAccountSpecificColName() string
	// SQL Table Schema
	GetTableSchema() string

	/*** Read/Get Queries ***/

	// Returns a query to get all accounts
	GetAllQuery(height int64) string
	// Returns a query to get the balance of an account at a specified height
	GetAccountAmountQuery(address string, height int64) string
	// Returns a query to select all accounts updated at a specified height
	GetAccountsUpdatedAtHeightQuery(height int64) string

	/*** Create/Insert Queries ***/

	// Returns a query to insert an account amount at a specified height
	InsertAccountQuery(address, amount string, height int64) string

	/*** Debug Queries Only ***/

	// Returns a query to clear all accounts
	ClearAllAccounts() string
}
