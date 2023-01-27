package types

func (s *Signature) ValidateBasic() Error {
	if s.Signature == nil {
		return ErrEmptySignature()
	}
	if s.PublicKey == nil {
		return ErrEmptyPublicKey()
	}
	return nil
}
