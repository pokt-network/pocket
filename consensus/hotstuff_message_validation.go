package consensus

import (
	"fmt"
	"pocket/consensus/types"
	"pocket/shared/crypto"
)

func (m *ConsensusModule) isMessagePartialSigValid(message *HotstuffMessage) bool {
	// Special case for development. No node will have ID 0.
	// TODO: Can this create a vulnerability?
	if message.Sender == 0 {
		return true
	}

	validator, ok := types.GetTestState().ValidatorMap[message.Sender]
	if !ok {
		m.nodeLog(fmt.Sprintf("[WARN] Trying to verify PartialSig from %d but it is not in the validator map.", message.Sender))
		return false
	}

	pubKey, err := crypto.NewPublicKey(validator.PublicKey)
	if err != nil {
		panic(err) // todo remove
		return false
	}
	if message.PartialSig != nil && !message.IsSignatureValid(pubKey, message.PartialSig) {
		m.nodeLogError(fmt.Sprintf("Partial signature on message is invalid. Sender: %d; Height: %d; Step: %d; Round: %d; SigHash: %s; BlockHash: %s; PubKey: %s", message.Sender, message.Height, message.Step, message.Round, message.PartialSig.HashString(), types.ProtoHash(message.Block), pubKey.String()))
		return false
	}
	return true
}

// TODO: Should this return an error or simply log every locally?
// TODO: Should this be part of the PaceMaker?
func (m *ConsensusModule) isValidProposal(message *HotstuffMessage) bool {
	// A nil QC implies successful commit or timeout. Not implementing CommitQC or TimeoutQC for now.
	if message.JustifyQC != nil && m.isQCValid(message.JustifyQC) {
		m.nodeLogError(fmt.Sprintf("[INVALID PROPOSAL] Quorum certificates on message is invalid: %+v", message))
		return false
	}

	lockedQC := m.LockedQC
	justifyQC := message.JustifyQC

	// Not locked.
	if lockedQC == nil {
		return true
	}

	// Safety. TODO: Implement `ExtendsFrom`
	if types.ProtoHash(lockedQC.Block) == types.ProtoHash(justifyQC.Block) { // && lockedQC.Block.ExtendsFrom(justifyQC.Block)
		return true
	}

	// Liveness check.
	if justifyQC.Height > lockedQC.Height || (justifyQC.Height == lockedQC.Height && justifyQC.Round > lockedQC.Round) {
		return true
	}

	return false
}
