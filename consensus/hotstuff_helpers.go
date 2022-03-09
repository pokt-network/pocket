package consensus

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/types"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

const ByzantineThreshold float64 = float64(2) / float64(3)

func (m *consensusModule) didReceiveEnoughMessageForStep(step types_consensus.HotstuffStep) (bool, string) {
	return m.isOptimisticThresholdMet(len(m.MessagePool[step]))
}

func (m *consensusModule) isOptimisticThresholdMet(n int) (bool, string) {
	valMap := types.GetTestState(nil).ValidatorMap
	return float64(n) > ByzantineThreshold*float64(len(valMap)), fmt.Sprintf("byzantine safety check: (%d > %.2f?)", n, ByzantineThreshold*float64(len(valMap)))
}

func (m *consensusModule) getQuorumCertificateForStep(step types_consensus.HotstuffStep) (*types_consensus.QuorumCertificate, error) {
	var pss []*types_consensus.PartialSignature
	for _, message := range m.MessagePool[step] {
		// TODO(olshansky): We're not validating that all the messages have the same height,
		// round and step when computing the ThresholdSignature. This can be fixed by making
		// the appropriate query to the persistence module. ADD TESTS.
		ps := message.GetPartialSignature()
		if ps == nil {
			m.nodeLog(fmt.Sprintf("[WARN] No partial signature found for step %s which should not happen...", StepToString[step]))
			continue
		}
		if ps.Signature == nil || len(ps.Address) == 0 {
			m.nodeLog(fmt.Sprintf("[WARN] Partial signature is incomplete for step %s which should not happen...", StepToString[step]))
			continue
		}
		pss = append(pss, message.GetPartialSignature())
	}

	if ok, reason := m.isOptimisticThresholdMet(len(pss)); !ok {
		return nil, fmt.Errorf("Did not receive enough partial signature; %s", reason)
	}
	thresholdSig, err := GetThresholdSignature(pss)
	if err != nil {
		return nil, err
	}

	return &types_consensus.QuorumCertificate{
		Height:             m.Height,
		Step:               step,
		Round:              m.Round,
		Block:              m.Block,
		ThresholdSignature: thresholdSig,
	}, nil
}

func (m *consensusModule) findHighQC(step types_consensus.HotstuffStep) (qc *types_consensus.QuorumCertificate) {
	for _, m := range m.MessagePool[step] {
		if m.GetQuorumCertificate() == nil {
			continue
		}
		if qc == nil || m.GetQuorumCertificate().Height > qc.Height {
			qc = m.GetQuorumCertificate()
		}
	}
	return
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
	for _, partialSig := range qc.ThresholdSignature.Signatures {
		validator, ok := valMap[partialSig.Address]
		if !ok {
			// TODO(olshansky): Remove this check. Even if we can't validate some partial signature, we could still meet byzantine safety.
			return false, fmt.Sprintf("Validator %d not found in the ValMap but a partial sig was signed by them.", m.ValToIdMap[partialSig.Address])
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without reserializing every time.
		if !isSignatureValid(msgToJustify, validator.PublicKey, partialSig.Signature) {
			return false, fmt.Sprintf("QC invalid because partial signature from the following node is invalid: %d\n", m.ValToIdMap[partialSig.Address])
		}
	}

	return true, "QC is valid"
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
