package consensus

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/types"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

const ByzantineThreshold float64 = float64(2) / float64(3)

func (m *consensusModule) didReceiveEnoughMessageForStep(step types_consensus.HotstuffStep) bool {
	return m.isOptimisticThresholdMet(len(m.MessagePool[step]))
}

func (m *consensusModule) isOptimisticThresholdMet(n int) bool {
	valMap := types.GetTestState(nil).ValidatorMap
	m.nodeLog(fmt.Sprintf("[DEBUG] Checking byzantine safety: %d > %.2f?", n, ByzantineThreshold*float64(len(valMap))))
	return float64(n) > ByzantineThreshold*float64(len(valMap))
}

func (m *consensusModule) getQCForStep(step types_consensus.HotstuffStep) (*types_consensus.QuorumCertificate, error) {
	var pss []*types_consensus.PartialSignature
	for _, message := range m.MessagePool[step] {
		// TODO(olshansky): We're not validating that all the messages have the same height,
		// round and step when computing the ThresholdSignature. This can be fixed by making
		// the appropriate query to the persistence module.
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

	if !m.isOptimisticThresholdMet(len(pss)) {
		return nil, fmt.Errorf("did not receive enough partial signature for Byzantine safety: %d/%d", len(pss), len(types.GetTestState(nil).ValidatorMap))
	}

	return &types_consensus.QuorumCertificate{
		Height:    m.Height,
		Step:      step,
		Round:     m.Round,
		Block:     m.Block,
		Signature: GetThresholdSignature(pss),
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

func (m *consensusModule) isQCValid(qc *types_consensus.QuorumCertificate) bool {
	if qc == nil {
		m.nodeLog("[WARN] Checking if a nil QC is valid...")
		return false
	}

	messageToJustify := QuorumCertificateToHotstuffMessage(qc)
	valMap := types.GetTestState(nil).ValidatorMap
	for _, partialSig := range qc.Signature.Signatures {
		validator, ok := valMap[partialSig.Address]
		if !ok {
			// TODO(olshansky): There is an optimization here where we can continue as long as
			// we still meet the byazantine minimum, but just fail fast for now.
			m.nodeLog(fmt.Sprintf("[WARN] Validator %d not found in the ValMap but a partial sig was signed by them.", m.ValToIdMap[partialSig.Address]))
			return false
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without reserializing every time.
		if !IsSignatureValid(messageToJustify, validator.PublicKey, partialSig.Signature) {
			m.nodeLog(fmt.Sprintf("[WARN] QC invalid because partial signature from the following node is invalid: %d\n", m.ValToIdMap[partialSig.Address]))
			return false
		}
	}
	return true
}
