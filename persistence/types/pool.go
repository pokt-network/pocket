package types

var _ ProtocolAccountSchema = &PoolSchema{}

type PoolSchema struct {
	baseProtocolAccountSchema
}

const (
	PoolTableName        = "pool"
	PoolHeightConstraint = "pool_create_height"
)

var Pool ProtocolAccountSchema = &PoolSchema{
	baseProtocolAccountSchema{
		tableName:              PoolTableName,
		accountSpecificColName: NameCol,
		heightConstraintName:   PoolHeightConstraint,
	},
}
