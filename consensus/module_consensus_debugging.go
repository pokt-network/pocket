package consensus

// All the code below is used for debugging & testing purposes only and should not be used in prod.
// TODO: Add debug/test tags to avoid accidental production usage.

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ConsensusDebugModule = &consensusModule{}

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

func (m *consensusModule) SetUtilityUnitOfWork(utilityUnitOfWork modules.UtilityUnitOfWork) {
	m.utilityUnitOfWork = utilityUnitOfWork
}

func (m *consensusModule) GetLeaderForView(height, round uint64, step uint8) uint64 {
	msg := &typesCons.HotstuffMessage{
		Height: height,
		Round:  round,
		Step:   typesCons.HotstuffStep(step),
	}
	leaderId, err := m.leaderElectionMod.ElectNextLeader(msg)
	if err != nil {
		return 0
	}
	return uint64(leaderId)
}
