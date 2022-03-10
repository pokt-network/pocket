package consensus

import (
	"fmt"
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

func CreateProposeMessage(
	m *consensusModule,
	step types_consensus.HotstuffStep,
	qc *types_consensus.QuorumCertificate,
) (*types_consensus.HotstuffMessage, error) {
	if m.Block == nil {
		return nil, fmt.Errorf("when a leader is trying to create a ProposeMessage, the block should never be nil")
	}

	msg := &types_consensus.HotstuffMessage{
		Type:          Propose,
		Height:        uint64(m.Height),
		Step:          step, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
		Round:         m.Round,
		Block:         m.Block,
		Justification: nil, // QC is set below if it is non-nil
	}

	if qc != nil {
		msg.Justification = &types_consensus.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	return msg, nil
}

func CreateVoteMessage(
	m *consensusModule,
	step types_consensus.HotstuffStep,
	block *types_consensus.BlockConsensusTemp,
) (*types_consensus.HotstuffMessage, error) {
	if block == nil {
		return nil, fmt.Errorf("replica should never vote on a nil block proposal")
	}

	msg := &types_consensus.HotstuffMessage{
		Type:          Vote,
		Height:        m.Height,
		Step:          step, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
		Round:         m.Round,
		Block:         block,
		Justification: nil, // signature is computed below
	}

	msg.Justification = &types_consensus.HotstuffMessage_PartialSignature{
		PartialSignature: &types_consensus.PartialSignature{
			Signature: getMessageSignature(msg, m.privateKey),
			Address:   m.privateKey.PublicKey().Address().String(),
		},
	}

	return msg, nil
}

func isSignatureValid(m *types_consensus.HotstuffMessage, pubKey crypto.PublicKey, signature []byte) bool {
	bytesToVerify, err := getSignableBytes(m)
	if err != nil {
		log.Println("[WARN] Error getting bytes to verify:", err)
		return false
	}
	return pubKey.VerifyBytes(bytesToVerify, signature)
}

func getMessageSignature(m *types_consensus.HotstuffMessage, privKey crypto.PrivateKey) []byte {
	bytesToSign, err := getSignableBytes(m)
	if err != nil {
		return nil
	}
	signature, err := privKey.Sign(bytesToSign)
	if err != nil {
		log.Fatalf("Error signing message: %v", err)
		return nil
	}
	return signature
}

func getSignableBytes(m *types_consensus.HotstuffMessage) ([]byte, error) {
	signableStruct := struct {
		Step   types_consensus.HotstuffStep `json:"step"`
		Height uint64                       `json:"height"`
		Round  uint64                       `json:"round"`
		Block  []byte                       `json:"block"`
	}{
		m.Step,
		m.Height,
		m.Round,
		types_consensus.ProtoMarshall(m.Block),
	}
	return types_consensus.GobEncode(signableStruct)
}
