package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/proto"
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

// CONSIDER! re-locating these functions to the state_sync module
// benefit of leaving them here is not expporting internal consensus module functions
// such as validateQuorumCertificate() and commitBlock
func (m *consensusModule) HandleGetBlockResponse(blockRes *typesCons.GetBlockResponse) error {

	serverNodePeerId := m.GetNodeAddress()
	clientPeerId := blockRes.PeerAddress

	fields := map[string]any{
		"currentHeight": blockRes.Block.BlockHeader.Height,
		"sender":        serverNodePeerId,
		"receiver":      clientPeerId,
	}

	m.logger.Info().Fields(fields).Msgf("Received GetBlockResponse: %s", blockRes)

	block := blockRes.Block
	lastPersistedBlockHeight := m.CurrentHeight() - 1

	if block.BlockHeader.Height <= m.CurrentHeight() {
		m.logger.Info().Msgf("Received block with height %d, but already at height %d, so not going to apply", block.BlockHeader.Height, lastPersistedBlockHeight)
		return nil
	}

	quorumCertBytes := block.BlockHeader.QuorumCertificate

	var qc *typesCons.QuorumCertificate
	err := proto.Unmarshal(quorumCertBytes, qc)
	if err != nil {
		return err
	}

	if err := m.validateQuorumCertificate(qc); err != nil {
		m.logger.Error().Err(err).Msg("Couldn't apply block, invalid QC")
		return err
	}

	if err := m.commitBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not commit block, invalid QC")
		return nil
	}

	return nil

}

func (m *consensusModule) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerId := consensusMod.GetNodeAddress()
	clientPeerId := metaDataRes.PeerAddress
	currentHeight := consensusMod.CurrentHeight()

	fields := map[string]any{
		"currentHeight": currentHeight,
		"sender":        serverNodePeerId,
		"receiver":      clientPeerId,
	}

	m.logger.Info().Fields(fields).Msgf("Received StateSyncMetadataResponse: %s", metaDataRes)

	//m.m.Lock()
	//defer m.m.Unlock()

	metaDataBuffer := m.stateSync.GetSyncMetadataBuffer()
	metaDataBuffer = append(metaDataBuffer, metaDataRes)
	m.stateSync.SetSyncMetadataBuffer(metaDataBuffer)

	return nil
}
