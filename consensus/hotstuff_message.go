package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

type HotstuffMessageType uint8

const (
	ProposeMessageType HotstuffMessageType = iota
	VoteMessageType
)

type HotstuffMessage struct {
	types_consensus.ConsensusMessage

	Type HotstuffMessageType

	Step   Step
	Height BlockHeight
	Round  Round
	Block  *types_consensus.BlockConsTemp

	// TODO: When moving to Protos, this should be a simple oneoff.
	JustifyQC  *QuorumCertificate // Non-nil from LEADER -> REPLICA; one of {HighQC, TimeoutQC, CommitQC}
	PartialSig Signature          // Non-nil from REPLICA -> LEADER; the replica signature over <height, step, block>.

	Sender types_consensus.NodeId
}

func (m *HotstuffMessage) GetType() types_consensus.ConsensusMessageType {
	return types_consensus.HotstuffConsensusMessage
}

func (m *HotstuffMessage) Encode() ([]byte, error) {
	bytes, err := types_consensus.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *HotstuffMessage) Decode(data []byte) error {
	err := types_consensus.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}

func CreateProposeMessage(m *consensusModule, step Step, qc *QuorumCertificate) (*HotstuffMessage, error) {
	if m.Block == nil {
		return nil, fmt.Errorf("If a leader is trying to create a ProposeMessage, the block should never be nil.")
	}

	message := &HotstuffMessage{
		Type: ProposeMessageType,

		Step:   step, // step is specified explicitly via interface to avoid ambiguity.
		Height: m.Height,
		Round:  m.Round,
		Block:  m.Block,

		JustifyQC: qc,

		Sender: m.NodeId,
	}
	return message, nil
}

func CreateVoteMessage(m *consensusModule, step Step, block *types_consensus.BlockConsTemp) (*HotstuffMessage, error) {
	if block == nil {
		return nil, fmt.Errorf("If a replica is trying to vote, the block should never be nil.")
	}

	message := &HotstuffMessage{
		Type: VoteMessageType,

		Step:   step, // step is specified explicitly via interface to avoid ambiguity.
		Height: m.Height,
		Round:  m.Round,
		Block:  block,

		Sender: m.NodeId,
	}
	message.PartialSig = message.getSignature(m.PrivateKey)

	return message, nil
}

func (m *HotstuffMessage) IsSignatureValid(pubKey crypto.PublicKey, sig Signature) bool {
	bytesToVerify, err := types_consensus.GobEncode(m.getSignableStruct())
	if err != nil {
		return false
	}
	return pubKey.VerifyBytes(bytesToVerify, sig)
}

func (m *HotstuffMessage) getSignableStruct() interface{} {
	return struct {
		Step   Step        `json:"step"`
		Height BlockHeight `json:"height"`
		Round  Round       `json:"round"`
		Block  []byte      `json:"block"`
	}{m.Step, m.Height, m.Round, types_consensus.ProtoMarshall(m.Block)}
}

func (m *HotstuffMessage) getSignature(privKey crypto.PrivateKey) Signature {
	bytesToSign, err := types_consensus.GobEncode(m.getSignableStruct())
	if err != nil {
		return nil
	}
	s, err := privKey.Sign(bytesToSign)
	if err != nil {
		panic(err) // remove
		return nil
	}
	sig := Signature(s)
	return sig
}
