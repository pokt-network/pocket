package types

import "fmt"

func (s *Signature) ValidateBasic() error {
	if s.Signature == nil {
		return fmt.Errorf("signature cannot be empty")
	}
	if s.PublicKey == nil {
		return fmt.Errorf("public key cannot be empty")
	}
	return nil
}
