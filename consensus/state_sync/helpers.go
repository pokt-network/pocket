package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Helper function for broadcasting state sync messages to the all peers known to the node:
//
//		sends metadata requests, via `metadataSyncLoop()` function
//	 	sends block requests, via `()` function
func (m *stateSync) broadcastStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, block_height uint64) error {
	m.logger.Info().Msg("📣 Broadcasting state sync message... 📣")

	currentHeight := m.bus.GetConsensusModule().CurrentHeight()

	validators, err := m.bus.GetConsensusModule().GetValidatorsAtHeight(currentHeight)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	// TODO: Use RainTree for this
	// IMPROVE: OPtimize so this is not O(n^2)
	for _, val := range validators {
		if err := m.SendStateSyncMessage(stateSyncMsg, cryptoPocket.AddressFromString(val.GetAddress()), block_height); err != nil {
			return err
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

func (m *stateSync) StateSyncLogHelper(receiverPeerAddress string) map[string]any {
	consensusMod := m.GetBus().GetConsensusModule()

	return map[string]any{
		"height":              consensusMod.CurrentHeight(),
		"senderPeerAddress":   consensusMod.GetNodeAddress(),
		"receiverPeerAddress": receiverPeerAddress,
	}
}
