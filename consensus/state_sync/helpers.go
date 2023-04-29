package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// SendStateSyncMessage sends a state sync message after converting to any proto, to the given peer
func (m *stateSync) sendStateSyncMessage(msg *typesCons.StateSyncMessage, dst cryptoPocket.Address) error {
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

func (m *stateSync) stateSyncLogHelper(receiverPeerAddress string) map[string]any {
	consensusMod := m.GetBus().GetConsensusModule()

	return map[string]any{
		"height":              consensusMod.CurrentHeight(),
		"senderPeerAddress":   consensusMod.GetNodeAddress(),
		"receiverPeerAddress": receiverPeerAddress,
	}
}

// func (m *stateSync) getAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse {
// 	minHeight, maxHeight := uint64(1), uint64(1)
// 	chanLen := len(m.metadataReceived)

// 	for i := 0; i < chanLen; i++ {
// 		metadata := <-m.metadataReceived
// 		if metadata.MaxHeight > maxHeight {
// 			maxHeight = metadata.MaxHeight
// 		}
// 		if metadata.MinHeight < minHeight {
// 			minHeight = metadata.MinHeight
// 		}
// 	}

// 	return &typesCons.StateSyncMetadataResponse{
// 		PeerAddress: "unused_aggregated_metadata_address",
// 		MinHeight:   minHeight,
// 		MaxHeight:   maxHeight,
// 	}
// }
