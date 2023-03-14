package types

import (
	"github.com/pokt-network/pocket/shared/pokterrors"
)

func (s *Signature) ValidateBasic() pokterrors.Error {
	if s.Signature == nil {
		return pokterrors.UtilityErrEmptySignature()
	}
	if s.PublicKey == nil {
		return pokterrors.UtilityErrEmptyPublicKey()
	}
	return nil
}
