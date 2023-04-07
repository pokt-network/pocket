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
		if !m.stateSync.IsServerModEnabled() {
			return fmt.Errorf("server module is not enabled")
		}
		return m.stateSync.HandleStateSyncMetadataRequest(stateSyncMessage.GetMetadataReq())
	case *typesCons.StateSyncMessage_MetadataRes:
		return m.HandleStateSyncMetadataResponse(stateSyncMessage.GetMetadataRes())
	case *typesCons.StateSyncMessage_GetBlockReq:
		m.logger.Info().Str("proto_type", "GetBlockRequest").Msg("Handling StateSyncMessage MetadataReq")
		if !m.stateSync.IsServerModEnabled() {
			return fmt.Errorf("server module is not enabled")
		}
		return m.stateSync.HandleGetBlockRequest(stateSyncMessage.GetGetBlockReq())
	case *typesCons.StateSyncMessage_GetBlockRes:
		return m.HandleGetBlockResponse(stateSyncMessage.GetGetBlockRes())
	default:
		return fmt.Errorf("unspecified state sync message type")
	}
}

// TODO(#352): Implement the business logic for this function
func (m *consensusModule) HandleGetBlockResponse(blockRes *typesCons.GetBlockResponse) error {
	m.logger.Info().Fields(m.statesSyncLogHelper(blockRes.PeerAddress)).Msgf("Received StateSync GetBlockResponse: %s", blockRes)

	if blockRes.Block.BlockHeader.GetQuorumCertificate() == nil {
		m.logger.Error().Err(typesCons.ErrNoQcInReceivedBlock).Msg(typesCons.DisregardBlock)
		return typesCons.ErrNoQcInReceivedBlock
	}

	return nil
}

// TODO(#352): Implement the business logic for this function
func (m *consensusModule) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	m.logger.Info().Fields(m.statesSyncLogHelper(metaDataRes.PeerAddress)).Msgf("Received StateSync MetadataResponse: %s", metaDataRes)

	return nil
}
