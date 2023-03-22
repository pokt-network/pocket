package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ConsensusDebugModule = &consensusModule{}

func (m *consensusModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	fmt.Println("OLSH2 - About to try to acquire lock")
	m.m.Lock()
	fmt.Println("OLSH2 - Acquired lock")
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

func (m *consensusModule) SetBlock(block *coreTypes.Block) {
	m.block = block
}

func (m *consensusModule) SetUtilityContext(utilityContext modules.UtilityContext) {
	m.utilityContext = utilityContext
}
