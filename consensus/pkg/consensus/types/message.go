package types

import (
	"bytes"
	"encoding/gob"

	"pocket/consensus/pkg/types"
)

type ConsensusMessageType string

const (
	HotstuffConsensusMessage  ConsensusMessageType = "Hotstuff"
	DKGConsensusMessage       ConsensusMessageType = "DKG"
	StateSyncConsensusMessage ConsensusMessageType = "StateSync"
	LeaderElectionMessage     ConsensusMessageType = "LeaderElection"
	DebugConsensusMessage     ConsensusMessageType = "Debug"
	TxWrapperMessageType      ConsensusMessageType = "Transaction"
)

type GenericConsensusMessage interface {
	GetType() ConsensusMessageType
	Decode(data []byte) error
	Encode() ([]byte, error)
}

type ConsensusMessage struct {
	// TODO: When moving to protobufs, this can be a one-off.
	Message GenericConsensusMessage
	Sender  types.NodeId
}

func EncodeConsensusMessage(message *ConsensusMessage) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(message); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func DecodeConsensusMessage(data []byte) (*ConsensusMessage, error) {
	var buff = bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)
	consensusMessage := &ConsensusMessage{}
	if err := dec.Decode(consensusMessage); err != nil {
		return nil, err
	}
	return consensusMessage, nil
}
