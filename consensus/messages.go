package consensus

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

func CreateProposeMessage(
	m *consensusModule,
	step typesCons.HotstuffStep, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
	qc *typesCons.QuorumCertificate,
) (*typesCons.HotstuffMessage, error) {
	if m.Block == nil {
		return nil, typesCons.ErrNilBlockProposal
	}

	msg := &typesCons.HotstuffMessage{
		Type:          Propose,
		Height:        m.Height,
		Step:          step,
		Round:         m.Round,
		Block:         m.Block,
		Justification: nil, // QC is set below if it is non-nil
	}

	// TODO(olshansky): Add unit tests for this
	if qc == nil && step != Prepare {
		return nil, typesCons.ErrNilQCProposal
	}

	// TODO(olshansky): Add unit tests for this
	if qc != nil { // QC may optionally be nil for NEWROUND steps when everything is progressing smoothly
		msg.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	return msg, nil
}

func CreateVoteMessage(
	m *consensusModule,
	step typesCons.HotstuffStep, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
	block *typesCons.BlockConsensusTemp,
) (*typesCons.HotstuffMessage, error) {
	if block == nil {
		return nil, typesCons.ErrNilBlockVote
	}

	msg := &typesCons.HotstuffMessage{
		Type:          Vote,
		Height:        m.Height,
		Step:          step,
		Round:         m.Round,
		Block:         block,
		Justification: nil, // signature is computed below
	}

	msg.Justification = &typesCons.HotstuffMessage_PartialSignature{
		PartialSignature: &typesCons.PartialSignature{
			Signature: getMessageSignature(msg, m.privateKey),
			Address:   m.privateKey.PublicKey().Address().String(),
		},
	}

	return msg, nil
}

// Returns a "partial" signature of the hotstuff message from one of the validators
func getMessageSignature(m *typesCons.HotstuffMessage, privKey crypto.PrivateKey) []byte {
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
func getSignableBytes(m *typesCons.HotstuffMessage) ([]byte, error) {
	msgToSign := &typesCons.HotstuffMessage{
		Height: m.Height,
		Step:   m.Step,
		Round:  m.Round,
		Block:  m.Block,
	}
	return proto.Marshal(msgToSign)
}
