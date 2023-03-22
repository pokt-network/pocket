package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *stateSync) SendStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, peerId cryptoPocket.Address, height uint64) error {
	anyMsg, err := anypb.New(stateSyncMsg)
	if err != nil {
		return err
	}

	m.logger.Info().Fields(m.logHelper(string(peerId))).Msg("Sending StateSync Message")
	return m.sendToPeer(anyMsg, peerId)
}

// Helper function for sending state sync messages
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
