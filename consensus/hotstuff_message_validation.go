package consensus

import (
	"fmt"
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) isPartialSignatureValid(msg *types_consensus.HotstuffMessage) (bool, string) {
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
		return false, fmt.Sprintf("Trying to verify PartialSignature from %d but it is not in the validator map.", m.ValAddrToIdMap[address])
	}

	pubKey := validator.PublicKey
	if isSignatureValid(msg, pubKey, msg.GetPartialSignature().Signature) {
		return true, "Partial signature is valid"
	}

	return false, fmt.Sprintf("Partial signature on message is invalid. Sender: %d; Height: %d; Step: %d; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s", m.ValAddrToIdMap[address], msg.Height, msg.Step, msg.Round, msg.GetPartialSignature().Signature, types_consensus.ProtoHash(msg.Block), pubKey.String())
}

func (m *consensusModule) isProposalValid(msg *types_consensus.HotstuffMessage) (bool, string) {
	if !(msg.Type == Propose && msg.Step == Prepare) {
		return false, "Proposal is not valid in the PREPARE step"
	}

	if valid, reason := m.isValidBlock(msg.Block); !valid {
		return false, fmt.Sprintf("Invalid block in message: %s", reason)
	}

	// TODO(discuss): A nil QC implies a successfull CommitQC or TimeoutQC, which have been omitted intentionally since
	// they are not needed for consensus validity. However, if a QC is specified, it must be valid.
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
		return false, "[TODO]: Do we want to set `m.LockedQC = nil` here or something else?"
	}

	return false, "[WARN] UNHANDLED PROPOSAL VALIDATION CHECK"
}

// TODO(olshansky): Check basic message metadata for validity (hash, size, etc)
func (m *consensusModule) isMessageBasicValid(message *types_consensus.HotstuffMessage) (bool, string) {
	return true, "basic message metadata is valid"
}

func (m *consensusModule) isQuorumCertificateValid(qc *types_consensus.QuorumCertificate) (bool, string) {
	if qc == nil {
		return false, "QC being validated is nil"
	}

	if qc.Block == nil {
		return false, "QC must contain a non nil block"
	}

	if qc.ThresholdSignature == nil || len(qc.ThresholdSignature.Signatures) == 0 {
		return false, "QC must contains a non nil threshold signature"
	}

	msgToJustify := qcToHotstuffMessage(qc)
	valMap := types.GetTestState(nil).ValidatorMap
	numValid := 0
	for _, partialSig := range qc.ThresholdSignature.Signatures {
		validator, ok := valMap[partialSig.Address]
		if !ok {
			m.nodeLog(fmt.Sprintf("[WARN] Validator %d not found in the ValMap but a partial sig was signed by them.", m.ValAddrToIdMap[partialSig.Address]))
			continue
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without re-serializing every time.
		if !isSignatureValid(msgToJustify, validator.PublicKey, partialSig.Signature) {
			m.nodeLog(fmt.Sprintf("[WARN] QC invalid because partial signature from the following node is invalid: %d", m.ValAddrToIdMap[partialSig.Address]))
			continue
		}
		numValid++
	}

	if ok, reason := m.isOptimisticThresholdMet(numValid); !ok {
		return false, fmt.Sprintf("QC invalid because optimistic threshold is not met: %s", reason)
	}

	return true, "QC is valid"
}

func isSignatureValid(m *types_consensus.HotstuffMessage, pubKey crypto.PublicKey, signature []byte) bool {
	bytesToVerify, err := getSignableBytes(m)
	if err != nil {
		log.Println("[WARN] Error getting bytes to verify:", err)
		return false
	}
	return pubKey.VerifyBytes(bytesToVerify, signature)
}

func qcToHotstuffMessage(qc *types_consensus.QuorumCertificate) *types_consensus.HotstuffMessage {
	return &types_consensus.HotstuffMessage{
		Height: qc.Height,
		Step:   qc.Step,
		Round:  qc.Round,
		Block:  qc.Block,
		Justification: &types_consensus.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		},
	}
}

func getThresholdSignature(
	partialSigs []*types_consensus.PartialSignature,
) (*types_consensus.ThresholdSignature, error) {
	thresholdSig := &types_consensus.ThresholdSignature{}
	thresholdSig.Signatures = make([]*types_consensus.PartialSignature, len(partialSigs))
	for i, parpartialSig := range partialSigs {
		thresholdSig.Signatures[i] = parpartialSig
	}
	return thresholdSig, nil
}
