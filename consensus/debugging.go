package consensus

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
)

func (m *consensusModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch debugMessage.Action {
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		m.resetToGenesis(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		m.printNodeState(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		m.triggerNextView(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		m.togglePacemakerManualMode(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_BLOCK_REQ:
		m.sendGetBlockStateSyncMessage(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_SEND_METADATA_REQ:
		m.sendGetMetadataStateSyncMessage(debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *consensusModule) GetNodeState() typesCons.ConsensusNodeState {
	leaderId := typesCons.NodeId(0)
	if m.leaderId != nil {
		leaderId = *m.leaderId
	}

	return typesCons.ConsensusNodeState{
		NodeId:   m.nodeId,
		Height:   m.height,
		Round:    uint8(m.round),
		Step:     uint8(m.step),
		IsLeader: m.IsLeader(),
		LeaderId: leaderId,
	}
}

func (m *consensusModule) resetToGenesis(_ *messaging.DebugMessage) {
	m.nodeLog(typesCons.DebugResetToGenesis)

	m.height = 0
	m.ResetForNewHeight()
	m.clearLeader()
	m.clearMessagesPool()
	m.GetBus().GetPersistenceModule().HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	})
	m.GetBus().GetPersistenceModule().Start() // reload genesis state
}

func (m *consensusModule) printNodeState(_ *messaging.DebugMessage) {
	state := m.GetNodeState()
	m.nodeLog(typesCons.DebugNodeState(state))
}

func (m *consensusModule) triggerNextView(_ *messaging.DebugMessage) {
	m.nodeLog(typesCons.DebugTriggerNextView)

	currentHeight := m.height
	currentStep := m.step
	if currentHeight == 0 || (currentStep == Decide && m.paceMaker.IsManualMode()) {
		m.paceMaker.NewHeight()
	} else {
		m.paceMaker.InterruptRound("manual trigger")
	}

	if m.paceMaker.IsManualMode() {
		m.paceMaker.ForceNextView()
	}
}

func (m *consensusModule) togglePacemakerManualMode(_ *messaging.DebugMessage) {
	newMode := !m.paceMaker.IsManualMode()
	if newMode {
		m.nodeLog(typesCons.DebugTogglePacemakerManualMode("MANUAL"))
	} else {
		m.nodeLog(typesCons.DebugTogglePacemakerManualMode("AUTOMATIC"))
	}
	m.paceMaker.SetManualMode(newMode)
}

// requests current block from all validators
func (m *consensusModule) sendGetBlockStateSyncMessage(_ *messaging.DebugMessage) {
	blockHeight := m.CurrentHeight() - 1
	peerAddress := m.GetNodeAddress()

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: peerAddress,
				Height:      blockHeight,
			},
		},
	}

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.nodeLogError(typesCons.ErrPersistenceGetAllValidators.Error(), err)
	}

	for _, val := range validators {
		if err := m.stateSync.SendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress()), blockHeight); err != nil {
			m.nodeLogError(typesCons.ErrBroadcastMessage.Error(), err)
		}
	}
}

// requests metadata from all validators
func (m *consensusModule) sendGetMetadataStateSyncMessage(_ *messaging.DebugMessage) {
	blockHeight := m.CurrentHeight() - 1
	peerAddress := m.GetNodeAddress()

	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: peerAddress,
			},
		},
	}

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.nodeLogError(typesCons.ErrPersistenceGetAllValidators.Error(), err)
	}

	for _, val := range validators {
		if err := m.stateSync.SendStateSyncMessage(stateSyncMetaDataReqMessage, cryptoPocket.AddressFromString(val.GetAddress()), blockHeight); err != nil {
			m.nodeLogError(typesCons.ErrBroadcastMessage.Error(), err)
		}
	}

}
