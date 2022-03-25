package types

import "github.com/pokt-network/pocket/shared/types"

const (
	DoubleSignEvidenceType = 1
)

// TODO NOTE: there's no signature validation on the vote because we are unsure the current mode of vote signing
// TODO *Needs to add signatures to vote structure*
func (v *Vote) ValidateBasic() types.Error {
	if err := ValidatePublicKey(v.PublicKey); err != nil {
		return err
	}
	if err := ValidateHash(v.BlockHash); err != nil {
		return err
	}
	if v.Height < 0 {
		return types.ErrInvalidBlockHeight()
	}
	if v.Type != DoubleSignEvidenceType {
		return types.ErrInvalidEvidenceType()
	}
	return nil
}
