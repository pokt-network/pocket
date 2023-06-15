package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/proto"
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

	// NB: Note that this is the only case that calls a function in the consensus module (not the state sync submodule) since
	// consensus is the one responsible for calling business logic to apply and commit the blocks. State sync listens for events
	// that are a result of it.
	case *typesCons.StateSyncMessage_GetBlockRes:
		m.logger.Info().Str("proto_type", "GetBlockResponse").Msg("Handling StateSyncMessage GetBlockResponse")
		go m.tryToApplyRequestedBlock(stateSyncMessage.GetGetBlockRes())
		return nil

	default:
		return fmt.Errorf("unspecified state sync message type")
	}
}

// tryToApplyRequestedBlock tries to commit the requested Block received from a peer.
// Intended to be called via a background goroutine.
// CLEANUP: Investigate whether this should be part of `Consensus` or part of `StateSync`
func (m *consensusModule) tryToApplyRequestedBlock(blockResponse *typesCons.GetBlockResponse) {
	logger := m.logger.With().Str("source", "tryToApplyRequestedBlock").Logger()

	// Retrieve the block we're about to try and apply
	block := blockResponse.Block
	if block == nil {
		logger.Error().Msg("Received nil block in GetBlockResponse")
		return
	}
	logger.Info().Msgf("Received new block at height %d.", block.BlockHeader.Height)

	// Check what the current latest committed block height is
	maxPersistedHeight, err := m.maxPersistedBlockHeight()
	if err != nil {
		logger.Err(err).Msg("couldn't query max persisted height")
		return
	}

	// Check if the block being synched is behind the current height
	if block.BlockHeader.Height <= maxPersistedHeight {
		logger.Debug().Msgf("Discarding block height %d, since node is ahead at height %d", block.BlockHeader.Height, maxPersistedHeight)
		return
	}

	// Check if the block being synched is ahead of the current height
	if block.BlockHeader.Height > m.CurrentHeight() {
		// IMPROVE: we need to store block responses that we are not yet ready to validate so we can validate them on a subsequent iteration of this loop
		logger.Info().Bool("TODO", true).Msgf("Received block at height %d, discarding as it is higher than the current height", block.BlockHeader.Height)
		return
	}

	// Perform basic validation on the block
	if err = m.basicValidateBlock(block); err != nil {
		logger.Err(err).Msg("failed to validate block")
		return
	}

	// Update the leader proposing the block
	// TECHDEBT: This ID logic could potentially be simplified in the future but needs a SPIKE
	leaderIdInt, err := m.getNodeIdFromNodeAddress(string(block.BlockHeader.ProposerAddress))
	if err != nil {
		m.logger.Error().Err(err).Msg("Could not get leader id from leader address")
		return
	}
	m.leaderId = typesCons.NewNodeId(leaderIdInt)

	// Prepare the utility UOW of work to apply a new block
	if err := m.refreshUtilityUnitOfWork(); err != nil {
		m.logger.Error().Err(err).Msg("Could not refresh utility context")
		return
	}

	// Try to apply the block by validating the transactions in the block
	if err := m.applyBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not apply block")
		return
	}

	// Try to commit the block to persistence
	if err := m.commitBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not commit block")
		return
	}
	logger.Info().Int64("height", int64(block.BlockHeader.Height)).Msgf("State sync committed block at height %d!", block.BlockHeader.Height)

	m.paceMaker.NewHeight()
}

// REFACTOR(#434): Once we consolidated NodeIds/PeerIds, this could potentially be removed
func (m *consensusModule) getNodeIdFromNodeAddress(peerId string) (uint64, error) {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.logger.Warn().Err(err).Msgf("Could not get validators at height %d when checking if peer %s is a validator", m.CurrentHeight(), peerId)
		return 0, fmt.Errorf("Could determine if peer %s is a validator or not: %w", peerId, err)
	}

	valAddrToIdMap := typesCons.NewActorMapper(validators).GetValAddrToIdMap()
	return uint64(valAddrToIdMap[peerId]), nil
}

// basicValidateBlock performs basic validation of the block, its metadata, signatures,
// but not the transactions themselves
func (m *consensusModule) basicValidateBlock(block *coreTypes.Block) error {
	blockHeader := block.BlockHeader
	qcBytes := blockHeader.GetQuorumCertificate()

	if qcBytes == nil {
		m.logger.Error().Err(typesCons.ErrNoQcInReceivedBlock).Msg(typesCons.DisregardBlock)
		return typesCons.ErrNoQcInReceivedBlock
	}

	qc := typesCons.QuorumCertificate{}
	if err := proto.Unmarshal(qcBytes, &qc); err != nil {
		return err
	}

	if err := m.validateQuorumCertificate(&qc); err != nil {
		m.logger.Error().Err(err).Msg("Couldn't apply block, invalid QC")
		return err
	}

	return nil
}
