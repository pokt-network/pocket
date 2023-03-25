package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO (#352): Implement this function, currently a placeholder.
// Helper function for broadcasting state sync messages to the all peers known to the node:
//
//		requests for metadata using the `periodicMetadataSynch()` function
//	 	requests for blocks using the `StartSynching()` function
func (m *stateSync) broadcastStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, height uint64) error {
	// TODO (#571): update with logger helper function
	m.logger.Info().Fields(
		map[string]any{
			"height": height,
			"nodeId": m.GetBus().GetConsensusModule().GetNodeId(),
		},
	).Msg("📣 Broadcasting state sync message... 📣")

	// TODO (#571) update, this is a placeholder
	_ = stateSyncMsg

	return nil
}

func (m *stateSync) SendStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, peerId cryptoPocket.Address, height uint64) error {
	anyMsg, err := anypb.New(stateSyncMsg)
	if err != nil {
		return err
	}

	m.logger.Info().Fields(m.logHelper(string(peerId))).Msg("Sending StateSync Message")
	return m.sendToPeer(anyMsg, peerId)
}

// Helper function for messages to the peers
func (m *stateSync) sendToPeer(msg *anypb.Any, peerId cryptoPocket.Address) error {
	if err := m.GetBus().GetP2PModule().Send(peerId, msg); err != nil {
		m.logger.Error().Msgf(typesCons.ErrSendMessage.Error(), err)
		return err
	}
	return nil
}

func (m *stateSync) logHelper(receiverPeerId string) map[string]any {
	consensusMod := m.GetBus().GetConsensusModule()

	return map[string]any{
		"height":         consensusMod.CurrentHeight(),
		"senderPeerId":   consensusMod.GetNodeAddress(),
		"receiverPeerId": receiverPeerId,
	}

}
