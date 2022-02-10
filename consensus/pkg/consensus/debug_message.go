package consensus

import (
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/shared"
)

type DebugMessageAction uint8

const (
	InterruptCurrentView DebugMessageAction = iota
	TriggerNextView
	TriggerDKG
	TogglePaceMakerManualMode
	ResetToGenesis
	PrintNodeState
)

type DebugMessage struct {
	consensus_types.GenericConsensusMessage

	Action DebugMessageAction
}

func (m *DebugMessage) GetType() consensus_types.ConsensusMessageType {
	return consensus_types.DebugConsensusMessage
}

func (m *DebugMessage) Encode() ([]byte, error) {
	bytes, err := shared.GobEncode(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *DebugMessage) Decode(data []byte) error {
	err := shared.GobDecode(data, m)
	if err != nil {
		return err
	}
	return nil
}
