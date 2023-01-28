package types

// No need for a Signature interface abstraction for the time being
var _ Validatable = &Signature{}

func (s *Signature) ValidateBasic() Error {
	if s.Signature == nil {
		return ErrEmptySignature()
	}
	if s.PublicKey == nil {
		return ErrEmptyPublicKey()
	}
	return nil
}
