package leader_election

import (
	"github.com/pokt-network/pocket/consensus/leader_election/sortition"
	"github.com/pokt-network/pocket/consensus/leader_election/vrf"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

type LeaderElectionMessageType uint8

const (
	VRFKeyBroadcast LeaderElectionMessageType = iota
	VRFProofBroadcast
)

type LeaderElectionMessage struct {
	types_consensus.ConsensusMessage

	Height uint64
	Round  uint64

	// TODO: This can be a one-off when we move to protobufs.
	Type     LeaderElectionMessageType
	KeyMsg   *LeaderElectionKeyBroadcastMessage
	ProofMsg *LeaderElectionProofBroadcastMessage

	Sender string // types_consensus.NodeId
}

type LeaderElectionKeyBroadcastMessage struct {
	VerificationKey vrf.VerificationKey
	VKStartHeight   uint64
	VKEndHeight     uint64
}

type LeaderElectionProofBroadcastMessage struct {
	VRFOut          vrf.VRFOutput
	VRFProof        vrf.VRFProof
	SortitionResult sortition.SortitionResult
}

func (m *LeaderElectionMessage) GetType() types_consensus.ConsensusMessageType {
	return types_consensus.ConsensusMessageType_CONSENSUS_UKNOWN_MESSAGE
}

func (m *LeaderElectionMessage) Encode() ([]byte, error) {
	bytes, err := types_consensus.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *LeaderElectionMessage) Decode(data []byte) error {
	err := types_consensus.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
