package types

var _ ProtocolAccountSchema = &PoolSchema{}

type PoolSchema struct {
	BaseProtocolAccountSchema
}

const (
	PoolTableName        = "pool"
	PoolHeightConstraint = "pool_create_height"
)

var Pool ProtocolAccountSchema = &PoolSchema{
	BaseProtocolAccountSchema{
		tableName:              PoolTableName,
		accountSpecificColName: NameCol,
		heightConstraintName:   PoolHeightConstraint,
	},
}
