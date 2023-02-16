package types

type Validatable interface {
	// ValidateBasic` is a stateless validation check that should encapsulate all
	// validations possible prior to interacting with the storage layer.
	ValidateBasic() Error
}
