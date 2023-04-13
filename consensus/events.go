package consensus

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/messaging"
)

// publishNewHeightEvent publishes a new height event to the bus so that other interested IntegratableModules can react to it if necessary
func (m *consensusModule) publishNewHeightEvent(height uint64) {
	newHeightEvent, err := messaging.PackMessage(&messaging.ConsensusNewHeightEvent{Height: height})
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Failed to pack consensus new height event")
	}
	m.GetBus().PublishEventToBus(newHeightEvent)

	fmt.Printf("Node address: %s, Event bus in consensus publishNewHeightEvent: %v\n", m.GetNodeAddress(), m.GetBus().GetEventBus())
}

// func (m *consensusModule) publishFSMEvent(msg *messaging.StateMachineTransitionEvent) {
// 	fsmEvent, err := messaging.PackMessage(msg)
// 	if err != nil {
// 		m.logger.Fatal().Err(err).Msg("Failed to pack consensus new height event")
// 	}
// 	m.GetBus().PublishEventToBus(fsmEvent)
// }
