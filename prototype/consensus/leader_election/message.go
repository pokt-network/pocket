package leader_election

import (
	"pocket/consensus/leader_election/sortition"
	"pocket/consensus/leader_election/vrf"
	types3 "pocket/consensus/types"
)

type LeaderElectionMessageType uint8

const (
	VRFKeyBroadcast LeaderElectionMessageType = iota
	VRFProofBroadcast
)

type LeaderElectionMessage struct {
	types3.ConsensusMessage

	Height types3.BlockHeight
	Round  types3.Round

	// TODO: This can be a one-off when we move to protobufs.
	Type     LeaderElectionMessageType
	KeyMsg   *LeaderElectionKeyBroadcastMessage
	ProofMsg *LeaderElectionProofBroadcastMessage

	Sender types3.NodeId
}

type LeaderElectionKeyBroadcastMessage struct {
	VerificationKey vrf.VerificationKey
	VKStartHeight   types3.BlockHeight
	VKEndHeight     types3.BlockHeight
}

type LeaderElectionProofBroadcastMessage struct {
	VRFOut          vrf.VRFOutput
	VRFProof        vrf.VRFProof
	SortitionResult sortition.SortitionResult
}

func (m *LeaderElectionMessage) GetType() types3.ConsensusMessageType {
	return types3.DKGConsensusMessage
}

func (m *LeaderElectionMessage) Encode() ([]byte, error) {
	bytes, err := types3.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *LeaderElectionMessage) Decode(data []byte) error {
	err := types3.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
