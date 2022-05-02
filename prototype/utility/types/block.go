package types

const (
	UnstakingStatus = 1
	StakedStatus    = 2

	DoubleSignEvidenceType = 1
)

func (x *Vote) ValidateBasic() Error {
	if err := ValidatePublicKey(x.PublicKey); err != nil {
		return err
	}
	if err := ValidateHash(x.BlockHash); err != nil {
		return err
	}
	if x.Height < 0 {
		return ErrInvalidBlockHeight()
	}
	if x.Type != DoubleSignEvidenceType {
		return ErrInvalidEvidenceType()
	}
	return nil
}
