package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *stateSync) BroadCastStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, height uint64) error {
	// anyMsg, err := anypb.New(stateSyncMsg)
	// if err != nil {
	// 	return err
	// }

	// validators, err := m. getValidatorsAtHeight(currentHeight)
	// if err != nil {
	// 	m.logger.Debug().Msgf(typesCons.ErrPersistenceGetAllValidators.Error(), err)
	// }

	// for _, val := range validators {
	// 	if m.GetNodeAddress() == val.GetAddress() {
	// 		continue
	// 	}
	// 	valAddress := cryptoPocket.AddressFromString(val.GetAddress())
	// 	if err := m.stateSync.SendStateSyncMessage(stateSyncGetBlockMessage, valAddress, requestHeight); err != nil {
	// 		m.logger.Error().Err(err).Str("proto_type", "GetBlockRequest").Msg("failed to send StateSyncMessage")
	// 	}
	// }

	// fields := map[string]any{
	// 	"height":     height,
	// 	"peerId":     peerId,
	// 	"proto_type": getMessageType(stateSyncMsg),
	// }

	// m.logger.Info().Fields(fields).Msg("Sending StateSync Message")
	// return m.sendToPeer(anyMsg, peerId)
	return nil
}

func (m *stateSync) SendStateSyncMessage(stateSyncMsg *typesCons.StateSyncMessage, peerId cryptoPocket.Address, height uint64) error {
	anyMsg, err := anypb.New(stateSyncMsg)
	if err != nil {
		return err
	}

	fields := map[string]any{
		"height":     height,
		"peerId":     peerId,
		"proto_type": getMessageType(stateSyncMsg),
	}

	m.logger.Info().Fields(fields).Msg("Sending StateSync Message")
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

func getMessageType(msg *typesCons.StateSyncMessage) string {
	switch msg.Message.(type) {
	case *typesCons.StateSyncMessage_MetadataReq:
		return "StateSyncMetadataRequest"
	case *typesCons.StateSyncMessage_MetadataRes:
		return "StateSyncMetadataResponse"
	case *typesCons.StateSyncMessage_GetBlockReq:
		return "GetBlockRequest"
	case *typesCons.StateSyncMessage_GetBlockRes:
		return "GetBlockResponse"
	default:
		return "Unknown"
	}
}
