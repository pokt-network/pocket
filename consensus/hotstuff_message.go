package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

func CreateProposeMessage(
	m *consensusModule,
	step types_consensus.HotstuffStep,
	qc *types_consensus.QuorumCertificate,
) (*types_consensus.HotstuffMessage, error) {
	if m.Block == nil {
		return nil, fmt.Errorf("if a leader is trying to create a ProposeMessage, the block should never be nil")
	}

	message := &types_consensus.HotstuffMessage{
		Type:   types_consensus.HotstuffMessageType_HOTSTUFF_MESAGE_PROPOSE,
		Step:   step, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
		Height: uint64(m.Height),
		Round:  m.Round,
		Block:  m.Block,
		Justification: &types_consensus.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		},
	}
	return message, nil
}

func CreateVoteMessage(
	m *consensusModule,
	step types_consensus.HotstuffStep,
	block *types_consensus.BlockConsensusTemp,
) (*types_consensus.HotstuffMessage, error) {
	if block == nil {
		return nil, fmt.Errorf("replica tring to vote on a nil block")
	}

	message := &types_consensus.HotstuffMessage{
		Type:          types_consensus.HotstuffMessageType_HOTSTUFF_MESSAGE_VOTE,
		Step:          step, // step can be taken from `m` but is specified explicitly via interface to avoid ambiguity
		Height:        m.Height,
		Round:         m.Round,
		Block:         block,
		Justification: nil, // signature is computed below
	}

	privKey := types.GetTestState(nil).PrivateKey // TODO(design): Is this where we should be storing/accessing the privateKey
	message.Justification = &types_consensus.HotstuffMessage_PartialSignature{
		PartialSignature: &types_consensus.PartialSignature{
			Signature: getHotstuffMessageSignature(message, privKey),
			Address:   privKey.PublicKey().Address().String(),
		},
	}

	return message, nil
}

func IsSignatureValid(m *types_consensus.HotstuffMessage, pubKey crypto.PublicKey, signature []byte) bool {
	bytesToVerify, err := types_consensus.GobEncode(getSignableStruct(m))
	if err != nil {
		return false
	}
	return pubKey.VerifyBytes(bytesToVerify, signature)
}

func getHotstuffMessageSignature(m *types_consensus.HotstuffMessage, privKey crypto.PrivateKey) []byte {
	bytesToSign, err := types_consensus.GobEncode(getSignableStruct(m))
	if err != nil {
		return nil
	}
	signature, err := privKey.Sign(bytesToSign)
	if err != nil {
		panic(err) // remove
		return nil
	}
	return signature
}

func getSignableStruct(m *types_consensus.HotstuffMessage) interface{} {
	return struct {
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
}
