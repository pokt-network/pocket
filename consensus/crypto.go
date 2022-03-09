package consensus

import (
	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

func GetThresholdSignature(
	partialSigs []*types_consensus.PartialSignature,
) (*types_consensus.ThresholdSignature, error) {
	thresholdSig := &types_consensus.ThresholdSignature{}
	thresholdSig.Signatures = make([]*types_consensus.PartialSignature, len(partialSigs))
	for i, parpartialSig := range partialSigs {
		thresholdSig.Signatures[i] = parpartialSig
	}
	return thresholdSig, nil
}
