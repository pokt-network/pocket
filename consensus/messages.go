package consensus

import (
	"fmt"
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

func CreateProposeMessage(
	m *consensusModule,
	step types_consensus.HotstuffStep, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
	qc *types_consensus.QuorumCertificate,
) (*types_consensus.HotstuffMessage, error) {
	if m.Block == nil {
		return nil, fmt.Errorf("when a leader is trying to create a ProposeMessage, the block should never be nil")
	}

	msg := &types_consensus.HotstuffMessage{
		Type:          Propose,
		Height:        m.Height,
		Step:          step,
		Round:         m.Round,
		Block:         m.Block,
		Justification: nil, // QC is set below if it is non-nil
	}

	// TODO(olshansky): Add unit tests for this
	if qc == nil && step != Prepare {
		return nil, fmt.Errorf("when creating a ProposeMessage for a step other than PREPARE, the qc should NEVER be nil")
	}

	// TODO(olshansky): Add unit tests for this
	if qc != nil { // QC may optionally be nil for NEWROUND steps when everything is progressing smoothly
		msg.Justification = &types_consensus.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	return msg, nil
}

func CreateVoteMessage(
	m *consensusModule,
	step types_consensus.HotstuffStep, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
	block *types_consensus.BlockConsensusTemp,
) (*types_consensus.HotstuffMessage, error) {
	if block == nil {
		return nil, fmt.Errorf("replica should never vote on a nil block proposal")
	}

	msg := &types_consensus.HotstuffMessage{
		Type:          Vote,
		Height:        m.Height,
		Step:          step,
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

// Returns a "partial" signature of the hotstuff message from one of the validators
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

// Signature should only be over a subset of the fields in a HotstuffMessage
func getSignableBytes(m *types_consensus.HotstuffMessage) ([]byte, error) {
	msgToSign := &types_consensus.HotstuffMessage{
		Height: m.Height,
		Step:   m.Step,
		Round:  m.Round,
		Block:  m.Block,
	}
	return proto.Marshal(msgToSign)
}
