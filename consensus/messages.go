package consensus

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

func CreateProposeMessage(
	height uint64,
	round uint64,
	step typesCons.HotstuffStep,
	block *typesCons.Block,
	qc *typesCons.QuorumCertificate,
) (*typesCons.HotstuffMessage, error) {
	if block == nil {
		return nil, typesCons.ErrNilBlockVote
	}

	msg := &typesCons.HotstuffMessage{
		Type:          Propose,
		Height:        height,
		Step:          step,
		Round:         round,
		Block:         block,
		Justification: nil, // QC is set below if it is non-nil
	}

	// TODO: Add unit tests for this
	if qc == nil && step != Prepare {
		return nil, typesCons.ErrNilQCProposal
	}

	// TODO: Add unit tests for this
	// QC may be nil during NEWROUND if following happy hotstuff path
	if qc != nil {
		msg.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	return msg, nil
}

func CreateVoteMessage(
	height uint64,
	round uint64,
	step typesCons.HotstuffStep,
	block *typesCons.Block,
	privKey crypto.PrivateKey, // used to sign the vote
) (*typesCons.HotstuffMessage, error) {
	if block == nil {
		return nil, typesCons.ErrNilBlockVote
	}

	msg := &typesCons.HotstuffMessage{
		Type:          Vote,
		Height:        height,
		Step:          step,
		Round:         round,
		Block:         block,
		Justification: nil, // signature is computed below
	}

	msg.Justification = &typesCons.HotstuffMessage_PartialSignature{
		PartialSignature: &typesCons.PartialSignature{
			Signature: getMessageSignature(msg, privKey),
			Address:   privKey.PublicKey().Address().String(),
		},
	}

	return msg, nil
}

// Returns "partial" signature of the hotstuff message from one of the validators.
// If there is an error signing the bytes, nil is returned instead.
func getMessageSignature(msg *typesCons.HotstuffMessage, privKey crypto.PrivateKey) []byte {
	bytesToSign, err := getSignableBytes(msg)
	if err != nil {
		log.Printf("[WARN] error getting bytes to sign: %v\n", err)
		return nil
	}

	signature, err := privKey.Sign(bytesToSign)
	if err != nil {
		log.Printf("[WARN] error signing message: %v\n", err)
		return nil
	}

	return signature
}

// Signature only over subset of fields in HotstuffMessage
// For reference, see section 4.3 of the the hotstuff whitepaper, partial signatures are
// computed over `tsignr(hm.type, m.viewNumber , m.nodei)`. https://arxiv.org/pdf/1803.05069.pdf
func getSignableBytes(msg *typesCons.HotstuffMessage) ([]byte, error) {
	msgToSign := &typesCons.HotstuffMessage{
		Height: msg.GetHeight(),
		Step:   msg.GetStep(),
		Round:  msg.GetRound(),
		Block:  msg.GetBlock(),
	}
	return codec.GetCodec().Marshal(msgToSign)
}
