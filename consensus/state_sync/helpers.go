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

	// TODO: Use RainTree for this
	// IMPROVE: OPtimize so this is not O(n^2)
	for _, val := range validators {
		validatorAddr := val.GetAddress()
		// DISCUSS_IN_THIS_COMMIT: You shouldn't need to do this check at the consensus module level - it's a P2P thin.
		if m.GetBus().GetConsensusModule().GetNodeAddress() != validatorAddr {
			if err := m.SendStateSyncMessage(stateSyncMsg, cryptoPocket.AddressFromString(val.GetAddress()), block_height); err != nil {
				return err
			}
		}
	}

	return nil
}

// SendStateSyncMessage sends a state sync message after converting to any proto, to the given peer
func (m *stateSync) SendStateSyncMessage(msg *typesCons.StateSyncMessage, dst cryptoPocket.Address, height uint64) error {
	anyMsg, err := anypb.New(msg)
	if err != nil {
		return err
	}

	if err := m.GetBus().GetP2PModule().Send(dst, anyMsg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
		return err
	}
	return nil
}

// TODO_IN_THIS_COMMIT: do not copy paste helpers; ditto below
func (m *stateSync) getValidatorsAtHeight(height uint64) ([]*coreTypes.Actor, error) {
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()

	return readCtx.GetAllValidators(int64(height))
}

func (m *stateSync) maximumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer readCtx.Release()

	maxHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}

func (m *stateSync) minimumPersistedBlockHeight() (uint64, error) {
	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return 0, err
	}
	defer readCtx.Release()

	maxHeight, err := readCtx.GetMinimumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}

func (m *stateSync) logHelper(receiverPeerAddress string) map[string]any {
	consensusMod := m.GetBus().GetConsensusModule()

	return map[string]any{
		"height":              consensusMod.CurrentHeight(),
		"senderPeerAddress":   consensusMod.GetNodeAddress(),
		"receiverPeerAddress": receiverPeerAddress,
	}
}
