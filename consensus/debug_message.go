package consensus

import (
	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

type DebugMessageAction uint8

const (
	InterruptCurrentView DebugMessageAction = iota
	TriggerNextView
	TriggerDKG
	TogglePaceMakerManualMode
	ResetToGenesis
	PrintNodeState
	SendTx
)

type DebugMessage struct {
	types_consensus.GenericConsensusMessage

	Action  DebugMessageAction
	Payload []byte
}

func (m *DebugMessage) GetType() types_consensus.ConsensusMessageType {
	return types_consensus.DebugConsensusMessage
}

func (m *DebugMessage) Encode() ([]byte, error) {
	bytes, err := types_consensus.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *DebugMessage) Decode(data []byte) error {
	err := types_consensus.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
