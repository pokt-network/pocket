package types

type Validatable interface {
	ValidatableBasic() Error
}
