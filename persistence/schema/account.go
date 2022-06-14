package schema

const (
	AccountTableName      = "account"
	PoolTableName         = "pool"
	AccountConstraintName = "account_create_height"
	PoolConstraintName    = "pool_create_height"
)

var (
	AccountTableSchema = AccTableSchema(Address, AccountConstraintName)
	PoolTableSchema    = AccTableSchema(Name, PoolConstraintName)
)

func GetAccountAmountQuery(address string, height int64) string {
	return SelectBalance(Address, address, height, AccountTableName)
}

func InsertAccountAmountQuery(address, amount string, height int64) string {
	return InsertAcc(Address, address, amount, height, AccountTableName, AccountConstraintName)
}

func GetPoolAmountQuery(name string, height int64) string {
	return SelectBalance(Name, name, height, PoolTableName)
}

func InsertPoolAmountQuery(name, amount string, height int64) string {
	return InsertAcc(Name, name, amount, height, PoolTableName, PoolConstraintName)
}
