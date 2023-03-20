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

// HandleGetBlockResponse handles the received block. It validates the block, quorum certificate and applies to its persistence
func (m *consensusModule) HandleGetBlockResponse(blockRes *typesCons.GetBlockResponse) error {
	m.logger.Info().Fields(m.logHelper(blockRes.PeerAddress)).Msgf("Received StateSync GetBlockResponse: %s", blockRes)

	block := blockRes.Block
	lastPersistedBlockHeight := m.CurrentHeight() - 1
	maxHeight, err := m.maximumPersistedBlockHeight()
	if err != nil {
		return err
	}

	m.logger.Info().Msgf("HandleGetBlockResponse, Starting, maxPersistedHeight is: %d", maxHeight)

	// checking if the received block is already persisted
	if block.BlockHeader.Height <= lastPersistedBlockHeight {
		m.logger.Info().Msgf("Received block with height %d, but already at height %d, so not going to apply", block.BlockHeader.Height, lastPersistedBlockHeight)
		return nil
	}

	blockHeader := block.BlockHeader
	qcBytes := blockHeader.GetQuorumCertificate()
	qc := typesCons.QuorumCertificate{}
	err = proto.Unmarshal(qcBytes, &qc)
	if err != nil {
		return err
	}

	m.logger.Info().Msg("HandleGetBlockResponse, validating Quroum Certificate")

	if err := m.validateQuorumCertificate(&qc); err != nil {
		m.logger.Error().Err(err).Msg("Couldn't apply block, invalid QC")
		return err
	}

	m.logger.Info().Msg("HandleGetBlockResponse, QC is valid, refreshing utility context")
	if err := m.refreshUtilityContext(); err != nil {
		m.logger.Error().Err(err).Msg("Could not refresh utility context")
		return err
	}

	m.logger.Info().Msg("HandleGetBlockResponse, committing the block")

	//m.m.Lock()
	//defer m.m.Unlock()
	if err := m.commitBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not commit block, invalid QC")
		return nil
	}

	m.paceMaker.NewHeight()

	maxHeight, err = m.maximumPersistedBlockHeight()
	if err != nil {
		return err
	}

	m.logger.Info().Msgf("HandleGetBlockResponse, Block is Committed, maxPersistedHeight is: %d, currentHeight is :%d", maxHeight, m.CurrentHeight())

	// Send current persisted block height to the state sync module
	m.stateSync.PersistedBlock(block.BlockHeader.Height)

	return nil

}

// HandleStateSyncMetadataResponse handles the received metadata. It updates state sync buffer
func (m *consensusModule) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	m.logger.Info().Fields(m.logHelper(metaDataRes.PeerAddress)).Msgf("Received StateSync MetadataResponse: %s", metaDataRes)

	metaDataBuffer := m.stateSync.GetStateSyncMetadataBuffer()
	metaDataBuffer = append(metaDataBuffer, metaDataRes)
	m.stateSync.SetStateSyncMetadataBuffer(metaDataBuffer)

	return nil
}
