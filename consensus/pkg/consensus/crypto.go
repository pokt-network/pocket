package consensus

import (
	"pocket/consensus/pkg/types"
	"pocket/shared"
	"pocket/shared/typespb"
)

type Signature []byte

// TODO: Until we figure out which library to use for threshold signatures,
// mimick the behaviour by looping over individual signatures.
type PartialSignature struct {
	Signature Signature
	NodeId    types.NodeId
}

type ThresholdSignature []PartialSignature

type QuorumCertificate struct {
	Height             BlockHeight
	Round              Round
	Step               Step
	Block              *typespb.Block
	ThresholdSignature ThresholdSignature
}

func GetThresholdSignature(partialSigs []*PartialSignature) *ThresholdSignature {
	thresholdSig := make(ThresholdSignature, len(partialSigs))
	for i, parpartialSig := range partialSigs {
		thresholdSig[i] = *parpartialSig
	}
	return &thresholdSig
}

func QCToHotstuffMessage(qc *QuorumCertificate) *HotstuffMessage {
	return &HotstuffMessage{
		Height: qc.Height,
		Round:  qc.Round,
		Step:   qc.Step,
		Block:  qc.Block,
	}
}

func (s *Signature) ToString() string {
	return shared.HexEncode(*s)
}

func (s *Signature) HashString() string {
	return shared.HashString(*s)
}
