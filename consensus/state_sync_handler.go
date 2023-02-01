package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) HandleStateSyncMessage(stateSyncMessageAny *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch stateSyncMessageAny.MessageName() {
	case StateSyncMessageContentType:
		msg, err := codec.GetCodec().FromAny(stateSyncMessageAny)
		if err != nil {
			return err
		}

		stateSyncMessage, ok := msg.(*typesCons.StateSyncMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to StateSyncMessage")
		}

		return m.handleStateSyncMessage(stateSyncMessage)
	default:
		return typesCons.ErrUnknownStateSyncMessageType(stateSyncMessageAny.MessageName())
	}

	return nil
}

func (m *consensusModule) handleStateSyncMessage(stateSyncMessage *typesCons.StateSyncMessage) error {
	switch stateSyncMessage.MsgType {
	case typesCons.StateSyncMessageType_STATE_SYNC_UNSPECIFIED:
		return fmt.Errorf("unspecified state sync message type")
	case typesCons.StateSyncMessageType_STATE_SYNC_METADATA_REQUEST:
		if !m.stateSync.IsServerModEnabled() {
			return fmt.Errorf("server module is not enabled")
		}
		return m.stateSync.HandleStateSyncMetadataRequest(stateSyncMessage.GetMetadataReq())
	case typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE:
		return m.stateSync.HandleStateSyncMetadataResponse(stateSyncMessage.GetMetadataRes())
	case typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_REQUEST:
		if !m.stateSync.IsServerModEnabled() {
			return fmt.Errorf("server module is not enabled")
		}
		return m.stateSync.HandleGetBlockRequest(stateSyncMessage.GetGetBlockReq())
	case typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_RESPONSE:
		return m.stateSync.HandleGetBlockResponse(stateSyncMessage.GetGetBlockRes())
	}

	return nil
}
