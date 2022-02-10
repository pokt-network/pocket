package consensus

import (
	"fmt"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/types"
	"pocket/consensus/pkg/types/typespb"
)

type HotstuffMessageType uint8

const (
	ProposeMessageType HotstuffMessageType = iota
	VoteMessageType
)

type HotstuffMessage struct {
	consensus_types.ConsensusMessage

	Type HotstuffMessageType

	Step   Step
	Height BlockHeight
	Round  Round
	Block  *typespb.Block

	// TODO: When moving to Protos, this should be a simple oneoff.
	JustifyQC  *QuorumCertificate // Non-nil from LEADER -> REPLICA; one of {HighQC, TimeoutQC, CommitQC}
	PartialSig Signature          // Non-nil from REPLICA -> LEADER; the replica signature over <height, step, block>.

	Sender types.NodeId
}

func (m *HotstuffMessage) GetType() consensus_types.ConsensusMessageType {
	return consensus_types.HotstuffConsensusMessage
}

func (m *HotstuffMessage) Encode() ([]byte, error) {
	bytes, err := shared.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *HotstuffMessage) Decode(data []byte) error {
	err := shared.GobDecode(data, m)
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

func CreateVoteMessage(m *consensusModule, step Step, block *typespb.Block) (*HotstuffMessage, error) {
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

func (m *HotstuffMessage) IsSignatureValid(pubKey types.PublicKey, sig Signature) bool {
	bytesToVerify, err := shared.GobEncode(m.getSignableStruct())
	if err != nil {
		return false
	}
	return pubKey.Verify(bytesToVerify, sig)
}

func (m *HotstuffMessage) getSignableStruct() interface{} {
	return struct {
		Step   Step        `json:"step"`
		Height BlockHeight `json:"height"`
		Round  Round       `json:"round"`
		Block  []byte      `json:"block"`
	}{m.Step, m.Height, m.Round, shared.ProtoMarshall(m.Block)}
}

func (m *HotstuffMessage) getSignature(privKey types.PrivateKey) Signature {
	bytesToSign, err := shared.GobEncode(m.getSignableStruct())
	if err != nil {
		return nil
	}
	sig := Signature(privKey.Sign(bytesToSign))
	return sig
}
