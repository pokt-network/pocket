package consensus

import (
	"github.com/pokt-network/pocket/shared/messaging"
)

// publishNewHeightEvent publishes a new height event to the bus so that other interested IntegratableModules can react to it if necessary
func (m *consensusModule) publishNewHeightEvent(height uint64) {
	newHeightEvent, err := messaging.PackMessage(&messaging.ConsensusNewHeightEvent{Height: height})
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Failed to pack consensus new height event")
	}
	m.GetBus().PublishEventToBus(newHeightEvent)
}

// publishStateSyncBlockCommittedEvent publishes a state_machine/module.goew state sync block committed event, so that state sync module can react to it
func (m *consensusModule) publishStateSyncBlockCommittedEvent(height uint64) {
	blockCommittedEvent := &messaging.StateSyncBlockCommittedEvent{
		Height: height,
	}
	stateSyncBlockCommittedEvent, err := messaging.PackMessage(blockCommittedEvent)
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Failed to pack state sync committed block event")
	}
	m.GetBus().PublishEventToBus(stateSyncBlockCommittedEvent)
}
