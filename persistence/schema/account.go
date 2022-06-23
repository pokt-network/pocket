package schema

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
