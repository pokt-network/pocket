package consensus

import (
	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

func GetThresholdSignature(
	partialSigs []*types_consensus.PartialSignature,
) *types_consensus.ThresholdSignature {
	thresholdSig := &types_consensus.ThresholdSignature{}
	thresholdSig.Signatures = make([]*types_consensus.PartialSignature, len(partialSigs))
	for i, parpartialSig := range partialSigs {
		thresholdSig.Signatures[i] = parpartialSig
	}
	return thresholdSig
}

func QuorumCertificateToHotstuffMessage(qc *types_consensus.QuorumCertificate) *types_consensus.HotstuffMessage {
	return &types_consensus.HotstuffMessage{
		Step:   qc.Step,
		Height: qc.Height,
		Round:  qc.Round,
		Block:  qc.Block,
		Justification: &types_consensus.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		},
	}
}
