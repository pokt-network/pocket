package types

var _ ProtocolAccountSchema = &AccountSchema{}

type AccountSchema struct {
	BaseProtocolAccountSchema
}

const (
	AccountTableName        = "account"
	AccountHeightConstraint = "account_create_height"
)

var Account ProtocolAccountSchema = &AccountSchema{
	BaseProtocolAccountSchema{
		tableName:              AccountTableName,
		accountSpecificColName: AddressCol,
		heightConstraintName:   AccountHeightConstraint,
	},
}
