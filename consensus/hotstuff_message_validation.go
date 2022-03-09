package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) isValidPartialSignature(msg *types_consensus.HotstuffMessage) (bool, string) {
	if msg.Step == NewRound {
		return true, "NewRound messages do not need a partial signature"
	}

	if msg.Type == Propose {
		return true, "Leader proposals do not need a partial signature"
	}

	if msg.GetPartialSignature() == nil {
		return false, "Partial signature cannot be nil"
	}

	if msg.GetPartialSignature().Signature == nil || len(msg.GetPartialSignature().Address) == 0 {
		return false, "Partial signature is either nil or source is not specified"
	}

	valMap := types.GetTestState(nil).ValidatorMap
	address := msg.GetPartialSignature().Address
	validator, ok := valMap[address]
	if !ok {
		return false, fmt.Sprintf("Trying to verify PartialSignature from %d but it is not in the validator map.", m.ValToIdMap[address])
	}

	pubKey := validator.PublicKey
	if isSignatureValid(msg, pubKey, msg.GetPartialSignature().Signature) {
		return true, "Partial signature is valid"
	}

	return false, fmt.Sprintf("Partial signature on message is invalid. Sender: %d; Height: %d; Step: %d; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s", m.ValToIdMap[address], msg.Height, msg.Step, msg.Round, msg.GetPartialSignature().Signature, types_consensus.ProtoHash(msg.Block), pubKey.String())
}

// TODO(olshansky): Should this be part of the PaceMaker?
func (m *consensusModule) isValidProposal(msg *types_consensus.HotstuffMessage) (bool, string) {
	if !(msg.Type == Propose && msg.Step == Prepare) {
		return false, "Proposal is not valid in the PREPARE step"
	}

	// TODO(discuss): Discuss the point below.
	// A nil QC implies a successfull CommitQC or TimeoutQC, which havebeen ommitted intentionally since they are
	// not needed for consensus validity. However, if a QC is specified, it must be valid.
	if msg.GetQuorumCertificate() != nil {
		if valid, reason := m.isQuorumCertificateValid(msg.GetQuorumCertificate()); !valid {
			return false, fmt.Sprintf("Proposal QC is invalid because: %s", reason)
		}
	}

	lockedQC := m.LockedQC
	justifyQC := msg.GetQuorumCertificate()

	// Safety: not locked
	if lockedQC == nil {
		return true, "Node is not locked on any QC"
	}

	// Safety: check the hash of the locked QC
	// TODO(olshansky): Extend implementation to adopt `ExtendsFrom` as described in the Hotstuff whitepaper.
	if types_consensus.ProtoHash(lockedQC.Block) == types_consensus.ProtoHash(justifyQC.Block) { // && lockedQC.Block.ExtendsFrom(justifyQC.Block)
		return true, "The ProposalQC block is the same as the LockedQC block"
	}

	// Liveness: node is locked on a QC from the past.
	if justifyQC.Height > lockedQC.Height || (justifyQC.Height == lockedQC.Height && justifyQC.Round > lockedQC.Round) {
		// m.LockedQC = nil
		return false, "[TODO]: Olshansky must discuss this case with Andrew"
	}

	return false, "UNHANDELED PROPOSAL VALIDATION CHECK"
}
