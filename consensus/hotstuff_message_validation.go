package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) isMessagePartialSigValid(message *types_consensus.HotstuffMessage) (bool, string) {
	if message.Step == NewRound {
		return true, "NewRound messages do not need a partial signature"
	}

	if message.Type == Propose {
		return true, "Leader proposals do not need a partial signature"
	}

	// TODO(olshansky): Remove this special case for debugging only.
	// if message.GetPartialSignature().Address == "DEBUG" {
	// 	return true, ""
	// }

	if message.GetPartialSignature() == nil {
		return false, "Partial signature is nil"
	}

	if message.GetPartialSignature().Signature == nil || message.GetPartialSignature().Address == "" {
		return false, "PartialSignature or internal attributes are nillis nil"
	}

	valMap := types.GetTestState(nil).ValidatorMap
	address := message.GetPartialSignature().Address

	validator, ok := valMap[address]
	if !ok {
		return false, fmt.Sprintf("Trying to verify PartialSig from %d but it is not in the validator map.", m.ValToIdMap[address])
	}

	pubKey := validator.PublicKey
	if !IsSignatureValid(message, pubKey, message.GetPartialSignature().Signature) {
		return false, fmt.Sprintf("Partial signature on message is invalid. Sender: %d; Height: %d; Step: %d; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s", m.ValToIdMap[address], message.Height, message.Step, message.Round, message.GetPartialSignature().Signature, types_consensus.ProtoHash(message.Block), pubKey.String())
	}

	return true, ""
}

// TODO(olshansky): Should this be part of the PaceMaker?
func (m *consensusModule) isValidProposal(message *types_consensus.HotstuffMessage) (bool, string) {
	// A nil QC implies successful CommitQC or TimeoutQC; these have not been implemented intentionally.
	if message.GetQuorumCertificate() != nil && m.isQCValid(message.GetQuorumCertificate()) {
		return false, fmt.Sprintf("Proposal QC is invalid: %+v", message)
	}

	lockedQC := m.LockedQC
	justifyQC := message.GetQuorumCertificate()

	// Not locked
	if lockedQC == nil {
		return true, "Node is not locked"
	}

	// Safety; TODO(olshansky): Implement `ExtendsFrom` as described in the Hotstuff whitepaper.
	if types_consensus.ProtoHash(lockedQC.Block) == types_consensus.ProtoHash(justifyQC.Block) { // && lockedQC.Block.ExtendsFrom(justifyQC.Block)
		return true, "ProposalQC is the same as the LockedQC"
	}

	// Liveness check
	if justifyQC.Height > lockedQC.Height || (justifyQC.Height == lockedQC.Height && justifyQC.Round > lockedQC.Round) {
		return true, "ProposalQC is ahead of the lockedQC, so the node is catching up"
	}

	return false, "UNHANDELED PROPOSAL VALIDATION CHECK"
}
