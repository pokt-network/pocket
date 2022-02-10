package leader_election

import (
	"pocket/consensus/pkg/consensus/leader_election/sortition"
	"pocket/consensus/pkg/consensus/leader_election/vrf"
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"
	"pocket/shared"
)

type LeaderElectionMessageType uint8

const (
	VRFKeyBroadcast LeaderElectionMessageType = iota
	VRFProofBroadcast
)

type LeaderElectionMessage struct {
	consensus_types.ConsensusMessage

	Height consensus_types.BlockHeight
	Round  consensus_types.Round

	// TODO: This can be a one-off when we move to protobufs.
	Type     LeaderElectionMessageType
	KeyMsg   *LeaderElectionKeyBroadcastMessage
	ProofMsg *LeaderElectionProofBroadcastMessage

	Sender types.NodeId
}

type LeaderElectionKeyBroadcastMessage struct {
	VerificationKey vrf.VerificationKey
	VKStartHeight   consensus_types.BlockHeight
	VKEndHeight     consensus_types.BlockHeight
}

type LeaderElectionProofBroadcastMessage struct {
	VRFOut          vrf.VRFOutput
	VRFProof        vrf.VRFProof
	SortitionResult sortition.SortitionResult
}

func (m *LeaderElectionMessage) GetType() consensus_types.ConsensusMessageType {
	return consensus_types.DKGConsensusMessage
}

func (m *LeaderElectionMessage) Encode() ([]byte, error) {
	bytes, err := shared.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *LeaderElectionMessage) Decode(data []byte) error {
	err := shared.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
