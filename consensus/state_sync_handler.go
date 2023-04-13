package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) HandleStateSyncMessage(stateSyncMessageAny *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	m.logger.Info().Msg("Handling StateSyncMessage")

	switch stateSyncMessageAny.MessageName() {
	case messaging.StateSyncMessageContentType:
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
}

func (m *consensusModule) handleStateSyncMessage(stateSyncMessage *typesCons.StateSyncMessage) error {
	switch stateSyncMessage.Message.(type) {
	case *typesCons.StateSyncMessage_MetadataReq:
		m.logger.Info().Str("proto_type", "MetadataRequest").Msg("Handling StateSyncMessage MetadataReq")
		if !m.serverModeEnabled {
			return fmt.Errorf("server module is not enabled")
		}
		return m.stateSync.HandleStateSyncMetadataRequest(stateSyncMessage.GetMetadataReq())
	case *typesCons.StateSyncMessage_MetadataRes:
		m.logger.Info().Str("proto_type", "MetadataResponse").Msg("Handling StateSyncMessage MetadataRes")
		m.metadataReceived <- stateSyncMessage.GetMetadataRes()
		return nil
	case *typesCons.StateSyncMessage_GetBlockReq:
		m.logger.Info().Str("proto_type", "GetBlockRequest").Msg("Handling StateSyncMessage GetBlockRequest")
		if !m.serverModeEnabled {
			return fmt.Errorf("server module is not enabled")
		}
		return m.stateSync.HandleGetBlockRequest(stateSyncMessage.GetGetBlockReq())
	case *typesCons.StateSyncMessage_GetBlockRes:
		m.logger.Info().Str("proto_type", "GetBlockResponse").Msg("Handling StateSyncMessage GetBlockResponse")
		m.blocksReceived <- stateSyncMessage.GetGetBlockRes()
		return nil
	default:
		return fmt.Errorf("unspecified state sync message type")
	}
}
