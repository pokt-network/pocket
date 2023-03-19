package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Helper function for broadcasting state sync messages to the all peers known to the node:
//
//		requests for metadata using the `periodicMetadataSynch()` function
//	 	requests for blocks using the `StartSynching()` function
func (m *stateSync) broadcastStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, block_height uint64) error {
	m.logger.Info().Msg("📣 Broadcasting state sync message... 📣")

	currentHeight := m.bus.GetConsensusModule().CurrentHeight()

	validators, err := m.getValidatorsAtHeight(currentHeight)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	for _, val := range validators {
		validatorAddr := val.GetAddress()
		if m.GetBus().GetConsensusModule().GetNodeAddress() != validatorAddr {
			if err := m.SendStateSyncMessage(stateSyncMsg, cryptoPocket.AddressFromString(val.GetAddress()), block_height); err != nil {
				return err
			}
		}

	}

	return nil
}

// SendStateSyncMessage sends a state sync message after converting to any proto, to the given peer
func (m *stateSync) SendStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, receiverPeerAddress cryptoPocket.Address, block_height uint64) error {
	consensusMod := m.GetBus().GetConsensusModule()

	anyMsg, err := anypb.New(stateSyncMsg)
	if err != nil {
		return err
	}

	// TODO (#571): update when #571 is merged
	fields := map[string]any{
		"height":              consensusMod.CurrentHeight(),
		"senderPeerAddress":   consensusMod.GetNodeAddress(),
		"receiverPeerAddress": receiverPeerAddress,
	}
	m.logger.Info().Fields(fields).Msg("Sending StateSync Message")

	//m.logger.Info().Msgf("NodeId: %d, NodeAddress: %s \n", m.GetBus().GetConsensusModule().GetNodeId(), receiverPeerAddress)

	if err := m.GetBus().GetP2PModule().Send(receiverPeerAddress, anyMsg); err != nil {
		m.logger.Error().Msgf(typesCons.ErrSendMessage.Error(), err)
		return err
	}
	return nil
}

func (m *stateSync) getValidatorsAtHeight(height uint64) ([]*coreTypes.Actor, error) {
	persistenceReadContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer persistenceReadContext.Close()

	return persistenceReadContext.GetAllValidators(int64(height))
}

func (m *stateSync) maximumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer persistenceContext.Close()

	maxHeight, err := persistenceContext.GetMaximumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}

func (m *stateSync) minimumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer persistenceContext.Close()

	maxHeight, err := persistenceContext.GetMinimumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}
