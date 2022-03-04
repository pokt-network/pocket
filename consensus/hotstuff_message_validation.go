package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) isMessagePartialSigValid(message *types_consensus.HotstuffMessage) (bool, string) {
	// Special case for development. No node will have ID 0.
	// TODO: Can this create a vulnerability?
	// if message.Sender == 0 {
	// 	return true
	// }

	if message.GetPartialSignature() == nil || message.GetPartialSignature().Signature == nil || message.GetPartialSignature().Address == "" {
		return false, "PartialSignature or internal attributes are nillis nil"
	}

	address := message.GetPartialSignature().Address
	nodeId, ok := m.ValToIdMap[address]
	if !ok {
		return false, fmt.Sprintf("Trying to get the nodeId for an address that's not in the validator mapping: %s", address)
	}

	valMap := types.GetTestState(nil).ValidatorMap
	validator, ok := valMap[address]
	if !ok {
		return false, fmt.Sprintf("[WARN] Trying to verify PartialSig from %d but it is not in the validator map.", m.ValToIdMap[address])
	}

	pubKey := validator.PublicKey
	if message.GetPartialSignature() != nil && !IsSignatureValid(message, pubKey, message.GetPartialSignature().Signature) {
		return false, fmt.Sprintf("Partial signature on message is invalid. Sender: %d; Height: %d; Step: %d; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s", nodeId, message.Height, message.Step, message.Round, message.GetPartialSignature().Signature, types_consensus.ProtoHash(message.Block), pubKey.String())
	}

	return true, ""
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
