package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO(#609): GetNodeState is currently exposed publicly so it can be accessed via reflection in tests. Refactor to use the test-only package and remove reflection
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

	// Reset Persistence to the genesis state
	if err := m.GetBus().GetPersistenceModule().HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	}); err != nil {
		return err
	}
	if err := m.GetBus().GetPersistenceModule().Start(); err != nil { // reload genesis state
		return err
	}

	// Reset Utility - must be done before consensus is restarted since it could affect the transactions in the next block
	m.GetBus().GetUtilityModule().GetMempool().Clear()

	// Restart consensus - must be done after the persistence module is cleared since it could affect the next elected leader
	m.ResetRound(true)
	m.SetHeight(0)

	return nil
}

func (m *consensusModule) printNodeState(_ *messaging.DebugMessage) {
	state := m.GetNodeState()
	m.logger.Debug().Fields(map[string]any{
		"step":   typesCons.StepToString[typesCons.HotstuffStep(state.Step)],
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

// sendGetBlockStateSyncMessage sends a messages to request specific blocks from peers
func (m *consensusModule) sendGetBlockStateSyncMessage(_ *messaging.DebugMessage) {
	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: m.GetNodeAddress(),
				Height:      m.CurrentHeight() - 1,
			},
		},
	}
	anyMsg, err := anypb.New(stateSyncGetBlockMessage)
	if err != nil {
		m.logger.Error().Err(err).Str("proto_type", "GetBlockRequest").Msg("failed to create StateSyncGetBlockMessage")
		return
	}
	if err := m.GetBus().GetP2PModule().Broadcast(anyMsg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrBroadcastMessage.Error())
		return
	}
}

// sendGetMetadataStateSyncMessage sends a message to request metadata from their peers
func (m *consensusModule) sendGetMetadataStateSyncMessage(_ *messaging.DebugMessage) {
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetNodeAddress(),
			},
		},
	}
	anyMsg, err := anypb.New(stateSyncMetaDataReqMessage)
	if err != nil {
		m.logger.Error().Err(err).Str("proto_type", "StateSyncMessage").Msg("failed to create StateSyncMetadataRequest")
		return
	}
	if m.GetBus().GetP2PModule().Broadcast(anyMsg) != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrBroadcastMessage.Error())
		return
	}
}
