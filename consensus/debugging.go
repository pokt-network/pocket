package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ modules.ConsensusDebugModule = &consensusModule{}
)

// Implementation of ConsensusDebugModule functions (i.e. SetHeight(), SetRound(), SetStep(), SetUtilityContext())
// exposed by the debug interface should only be used for testing purposes.

func (m *consensusModule) SetHeight(height uint64) {
	m.height = height
	m.publishNewHeightEvent(height)
}

func (m *consensusModule) SetRound(round uint64) {
	m.round = round
}

func (m *consensusModule) SetStep(step uint8) {
	m.step = typesCons.HotstuffStep(step)
}

func (m *consensusModule) SetBlock(block *coreTypes.Block) {
	m.block = block
}

func (m *consensusModule) SetUtilityContext(utilityContext modules.UtilityContext) {
	m.utilityContext = utilityContext
}

func (m *consensusModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch debugMessage.Action {
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		if err := m.resetToGenesis(debugMessage); err != nil {
			return err
		}
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
		m.logger.Debug().Msgf("Debug message: %s", debugMessage.Message)
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

func (m *consensusModule) resetToGenesis(_ *messaging.DebugMessage) error {
	m.logger.Debug().Msg(typesCons.DebugResetToGenesis)

	m.SetHeight(0)
	m.ResetForNewHeight()
	m.clearLeader()
	m.clearMessagesPool()
	m.GetBus().GetUtilityModule().GetMempool().Clear()
	if err := m.GetBus().GetPersistenceModule().HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	}); err != nil {
		return err
	}
	if err := m.GetBus().GetPersistenceModule().Start(); err != nil { // reload genesis state
		return err
	}
	return nil
}

func (m *consensusModule) printNodeState(_ *messaging.DebugMessage) {
	state := m.GetNodeState()
	m.logger.Debug().
		Fields(map[string]any{
			"step":   state.Step,
			"height": state.Height,
			"round":  state.Round,
		}).Msg("Node state")
}

func (m *consensusModule) triggerNextView(_ *messaging.DebugMessage) {
	m.logger.Debug().Msg(typesCons.DebugTriggerNextView)

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
		m.logger.Debug().Str("pacemaker_mode", "MANUAL").Msg("Toggle pacemaker to MANUAL mode")
	} else {
		m.logger.Debug().Str("pacemaker_mode", "AUTOMATIC").Msg("Toggle pacemaker to AUTOMATIC mode")
	}
	m.paceMaker.SetManualMode(newMode)
}

// requests current block from all validators
func (m *consensusModule) sendGetBlockStateSyncMessage(_ *messaging.DebugMessage) {
	currentHeight := m.CurrentHeight()
	requestHeight := currentHeight - 1
	peerAddress := m.GetNodeAddress()

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: peerAddress,
				Height:      requestHeight,
			},
		},
	}

	validators, err := m.getValidatorsAtHeight(currentHeight)
	if err != nil {
		m.logger.Debug().Msgf(typesCons.ErrPersistenceGetAllValidators.Error(), err)
	}

	for _, val := range validators {
		if m.GetNodeAddress() == val.GetAddress() {
			continue
		}
		valAddress := cryptoPocket.AddressFromString(val.GetAddress())
		if err := m.stateSync.SendStateSyncMessage(stateSyncGetBlockMessage, valAddress, requestHeight); err != nil {
			m.logger.Error().Err(err).Str("proto_type", "GetBlockRequest").Msg("failed to send StateSyncMessage")
		}
	}
}

// requests metadata from all validators
func (m *consensusModule) sendGetMetadataStateSyncMessage(_ *messaging.DebugMessage) {
	currentHeight := m.CurrentHeight()
	requestHeight := currentHeight - 1
	peerAddress := m.GetNodeAddress()

	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: peerAddress,
			},
		},
	}

	validators, err := m.getValidatorsAtHeight(currentHeight)
	if err != nil {
		m.logger.Debug().Msgf(typesCons.ErrPersistenceGetAllValidators.Error(), err)
	}

	for _, val := range validators {
		if m.GetNodeAddress() != val.GetAddress() {
			continue
		}
		valAddress := cryptoPocket.AddressFromString(val.GetAddress())
		if err := m.stateSync.SendStateSyncMessage(stateSyncMetaDataReqMessage, valAddress, requestHeight); err != nil {
			m.logger.Error().Err(err).Str("proto_type", "StateSyncMetadataRequest").Msg("failed to send StateSyncMessage")
		}
	}

}

// Implementations of the type PaceMakerAccessModule interface
//
//	SetHeight, SetRound, SetStep are implemented for ConsensusDebugModule
func (m *consensusModule) ClearLeaderMessagesPool() {
	m.clearLeader()
	m.clearMessagesPool()
}
