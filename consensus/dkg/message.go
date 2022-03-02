//go:build test
// +build test

package dkg

import (
	types_consensus "github.com/pokt-network/pocket/consensus/types"
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
	types_consensus.ConsensusMessage

	Round       DKGRound // Specifies the round that this messages serves as input to. For example, the output of bcast from Round 1 will have DKGRound2.
	MessageType DKGMessageType
	MessageData []byte

	Sender    types_consensus.NodeId
	Recipient *types_consensus.NodeId // nil if broadcast
}

func (m *DKGMessage) GetType() types_consensus.ConsensusMessageType {
	return types_consensus.DKGConsensusMessage
}

func (m *DKGMessage) Encode() ([]byte, error) {
	bytes, err := types_consensus.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *DKGMessage) Decode(data []byte) error {
	err := types_consensus.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
