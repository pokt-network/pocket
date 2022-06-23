package schema

const (
	AccountsTableName        = "accounts"
	AccountsHeightConstraint = "accounts_create_height"

	/*
		From the utility specification:
			A ModulePool is a particular type that though similar in structure to an Account, the
			functionality of each is quite specialized to its use case. These pools are maintained by
			the protocol and are completely autonomous, owned by no actor on the network. Unlike Accounts,
			tokens are able to be directly minted to and burned from ModulePools. Examples of ModuleAccounts
			include StakingPools and the FeeCollector
	*/
	PoolsTableName        = "pool"
	PoolsHeightConstraint = "pools_create_height"
)

var (
	AccountsTableSchema = AccountTableSchema(AddressCol, AccountsHeightConstraint)
	PoolsTableSchema    = AccountTableSchema(NameCol, PoolsHeightConstraint)
)

func GetAccountAmountQuery(address string, height int64) string {
	return SelectBalance(AddressCol, address, height, AccountsTableName)
}

func InsertAccountAmountQuery(address, amount string, height int64) string {
	return InsertAcc(AddressCol, address, amount, height, AccountsTableName, AccountsHeightConstraint)
}

func GetPoolAmountQuery(name string, height int64) string {
	return SelectBalance(NameCol, name, height, PoolsTableName)
}

func InsertPoolAmountQuery(name, amount string, height int64) string {
	return InsertAcc(NameCol, name, amount, height, PoolsTableName, PoolsHeightConstraint)
}
