package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) HandleStateSyncMessage(stateSyncMessageAny *anypb.Any) error {
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

	case messaging.StateSyncBlockCommittedEventType:
		return m.stateSync.HandleStateSyncBlockCommittedEvent(stateSyncMessageAny)

	default:
		return typesCons.ErrUnknownStateSyncMessageType(stateSyncMessageAny.MessageName())
	}
}

func (m *consensusModule) handleStateSyncMessage(stateSyncMessage *typesCons.StateSyncMessage) error {
	switch stateSyncMessage.Message.(type) {

	case *typesCons.StateSyncMessage_MetadataReq:
		m.logger.Info().Str("proto_type", "MetadataRequest").Msg("Handling StateSyncMessage MetadataReq")
		if !m.consCfg.ServerModeEnabled {
			m.logger.Warn().Msg("Node's server module is not enabled")
			return nil
		}
		go m.stateSync.HandleStateSyncMetadataRequest(stateSyncMessage.GetMetadataReq())
		return nil

	case *typesCons.StateSyncMessage_GetBlockReq:
		m.logger.Info().Str("proto_type", "GetBlockRequest").Msg("Handling StateSyncMessage GetBlockRequest")
		if !m.consCfg.ServerModeEnabled {
			m.logger.Warn().Msg("Node's server module is not enabled")
			return nil
		}
		go m.stateSync.HandleGetBlockRequest(stateSyncMessage.GetGetBlockReq())
		return nil

	case *typesCons.StateSyncMessage_MetadataRes:
		m.logger.Info().Str("proto_type", "MetadataResponse").Msg("Handling StateSyncMessage MetadataRes")
		go m.stateSync.HandleStateSyncMetadataResponse(stateSyncMessage.GetMetadataRes())
		return nil

	case *typesCons.StateSyncMessage_GetBlockRes:
		m.logger.Info().Str("proto_type", "GetBlockResponse").Msg("Handling StateSyncMessage GetBlockResponse")
		go m.tryToApplyRequestedBlock(stateSyncMessage.GetGetBlockRes())
		return nil

	default:
		return fmt.Errorf("unspecified state sync message type")
	}
}
