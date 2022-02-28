package consensus

import (
	consensus_types "pocket/consensus/types"
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
	consensus_types.GenericConsensusMessage

	Action  DebugMessageAction
	Payload []byte
}

func (m *DebugMessage) GetType() consensus_types.ConsensusMessageType {
	return consensus_types.DebugConsensusMessage
}

func (m *DebugMessage) Encode() ([]byte, error) {
	bytes, err := consensus_types.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *DebugMessage) Decode(data []byte) error {
	err := consensus_types.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
