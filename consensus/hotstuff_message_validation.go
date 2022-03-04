package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) isMessagePartialSigValid(message *types_consensus.HotstuffMessage) bool {
	// Special case for development. No node will have ID 0.
	// TODO: Can this create a vulnerability?
	// if message.Sender == 0 {
	// 	return true
	// }

	address := message.GetPartialSignature().Address
	nodeId, ok := m.NodeIdMap[address]
	if !ok {
		m.nodeLog(fmt.Sprintf("[WARN] Trying to get the nodeId for an address that's not in the validator mapping: %s", address))
		return false
	}

	valMap := types.GetTestState(nil).ValidatorMap
	validator, ok := valMap[address]
	if !ok {
		m.nodeLog(fmt.Sprintf("[WARN] Trying to verify PartialSig from %d but it is not in the validator map.", m.NodeIdMap[address]))
		return false
	}

	pubKey := validator.PublicKey
	if message.GetPartialSignature() != nil && !IsSignatureValid(message, pubKey, message.GetPartialSignature().Signature) {
		m.nodeLogError(fmt.Sprintf("Partial signature on message is invalid. Sender: %d; Height: %d; Step: %d; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s", nodeId, message.Height, message.Step, message.Round, message.GetPartialSignature().Signature, types_consensus.ProtoHash(message.Block), pubKey.String()), nil)
		return false
	}
	return true
}

// TODO: Should this return an error or simply log every locally?
// TODO: Should this be part of the PaceMaker?
func (m *consensusModule) isValidProposal(message *types_consensus.HotstuffMessage) bool {
	// A nil QC implies successful commit or timeout. Not implementing CommitQC or TimeoutQC for now.
	if message.GetQuorumCertificate() != nil && m.isQCValid(message.GetQuorumCertificate()) {
		m.nodeLogError(fmt.Sprintf("[INVALID PROPOSAL] Quorum certificates on message is invalid: %+v", message), nil)
		return false
	}

	lockedQC := m.LockedQC
	justifyQC := message.GetQuorumCertificate()

	// Not locked.
	if lockedQC == nil {
		return true
	}

	// Safety. TODO: Implement `ExtendsFrom`
	if types_consensus.ProtoHash(lockedQC.Block) == types_consensus.ProtoHash(justifyQC.Block) { // && lockedQC.Block.ExtendsFrom(justifyQC.Block)
		return true
	}

	// Liveness check.
	if justifyQC.Height > lockedQC.Height || (justifyQC.Height == lockedQC.Height && justifyQC.Round > lockedQC.Round) {
		return true
	}

	return false
}
