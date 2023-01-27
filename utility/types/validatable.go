package types

type Validatable interface {
	ValidateBasic() Error
}
