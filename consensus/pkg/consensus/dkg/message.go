package dkg

import (
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"
	"pocket/shared"
)

type DKGMessageType string

type DKGRound uint8

const (
	// Distributed Key Generation
	DKGRound1 DKGRound = iota
	DKGRound2
	DKGRound3
	DKGRound4
)

const (
	DKGBroadcast DKGMessageType = "gennaro.Round1Bcast"
	DKGP2PSend   DKGMessageType = "gennaro.Round1P2PSend"
)

type DKGMessage struct {
	consensus_types.ConsensusMessage

	Round       DKGRound // Specifies the round that this messages serves as input to. For example, the output of bcast from Round 1 will have DKGRound2.
	MessageType DKGMessageType
	MessageData []byte

	Sender    types.NodeId
	Recipient *types.NodeId // nil if broadcast
}

func (m *DKGMessage) GetType() consensus_types.ConsensusMessageType {
	return consensus_types.DKGConsensusMessage
}

func (m *DKGMessage) Encode() ([]byte, error) {
	bytes, err := shared.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *DKGMessage) Decode(data []byte) error {
	err := shared.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
