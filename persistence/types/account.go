package types

var _ ProtocolAccountSchema = &AccountSchema{}

type AccountSchema struct {
	baseProtocolAccountSchema
}

const (
	AccountTableName        = "account"
	AccountHeightConstraint = "account_create_height"
)

var Account ProtocolAccountSchema = &AccountSchema{
	baseProtocolAccountSchema{
		tableName:              AccountTableName,
		accountSpecificColName: AddressCol,
		heightConstraintName:   AccountHeightConstraint,
	},
}
