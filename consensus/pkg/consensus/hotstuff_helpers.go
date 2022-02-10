package consensus

import (
	"fmt"

	"pocket/shared"
)

const ByzantineThreshold float64 = float64(2) / float64(3)

func (m *consensusModule) didReceiveEnoughMessageForStep(step Step) bool {
	return m.isOptimisticThresholdMet(len(m.MessagePool[step]))
}

func (m *consensusModule) isOptimisticThresholdMet(n int) bool {
	valMap := shared.GetPocketState().ValidatorMap
	m.nodeLog(fmt.Sprintf("[DEBUG] Checking byzantine safety: %d > %.2f?", n, ByzantineThreshold*float64(len(valMap))))
	return float64(n) > ByzantineThreshold*float64(len(valMap))
}

func (m *consensusModule) getQCForStep(step Step) (*QuorumCertificate, error) {
	var pss []*PartialSignature
	for _, message := range m.MessagePool[step] {
		// TODO: We're not validating that all the messages have the same height, round and step when computing the TS.
		// This can be fixed by making the appropriate query to the persistence m.
		if message.PartialSig == nil {
			m.nodeLog(fmt.Sprintf("[WARN] No partial signature found for step %s from node %d which should not happem...", StepToString[step], message.Sender))
			continue
		}
		ps := &PartialSignature{
			NodeId:    message.Sender,
			Signature: message.PartialSig,
		}
		pss = append(pss, ps)
	}

	if !m.isOptimisticThresholdMet(len(pss)) {
		return nil, fmt.Errorf("did not receive enough partial signature for Byzantine safety: %d/%d", len(pss), len(shared.GetPocketState().ValidatorMap))
	}

	return &QuorumCertificate{
		Step:               step,
		Round:              m.Round,
		Height:             m.Height,
		Block:              m.Block,
		ThresholdSignature: *GetThresholdSignature(pss),
	}, nil
}

func (m *consensusModule) findHighQC(step Step) (qc *QuorumCertificate) {
	for _, m := range m.MessagePool[step] {
		if m.JustifyQC == nil {
			continue
		}
		if qc == nil || m.JustifyQC.Height > qc.Height {
			qc = m.JustifyQC
		}
	}
	return
}

func (m *consensusModule) isQCValid(qc *QuorumCertificate) bool {
	if qc == nil {
		m.nodeLog("[WARN] Checking if a nil QC is valid...")
		return false
	}

	messageToJustify := QCToHotstuffMessage(qc)
	valMap := shared.GetPocketState().ValidatorMap
	for _, partialSig := range qc.ThresholdSignature {
		validator, ok := valMap[partialSig.NodeId]
		if !ok {
			// TODO: There is an optimization here where we can continue as long as we still
			// meet the byazantine minimum, but just fail fast for now.
			m.nodeLog(fmt.Sprintf("[WARN] Validator %d not found in the ValMap but a partial sig was signed by them.", partialSig.NodeId))
			return false
		}
		// TODO: Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without reserializing every time.
		if !messageToJustify.IsSignatureValid(validator.PublicKey, partialSig.Signature) {
			m.nodeLog(fmt.Sprintf("[WARN] QC invalid because partial signature from the following node is invalid: %d\n", partialSig.NodeId))
			return false
		}
	}
	return true
}
