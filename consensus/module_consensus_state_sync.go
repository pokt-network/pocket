package consensus

import (
	"context"
	"fmt"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO: Make this configurable in StateSyncConfig
const metadataSyncPeriod = 45 * time.Second

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

// blockApplicationLoop commits the blocks received from the `blocksResponsesReceived“ channel.
// It is intended to be run as a background process via `go blockApplicationLoop()`
func (m *consensusModule) blockApplicationLoop() {
	logger := m.logger.With().Str("source", "blockApplicationLoop").Logger()

	// Blocks until m.blocksResponsesReceived is closed
	for blockResponse := range m.blocksResponsesReceived {
		block := blockResponse.Block
		logger.Info().Msgf("Received new block at height %d.", block.BlockHeader.Height)

		// Check what the current latest committed block height is
		maxPersistedHeight, err := m.maxPersistedBlockHeight()
		if err != nil {
			logger.Err(err).Msg("couldn't query max persisted height")
			continue
		}

		// Check if the block being synched is behind the current height
		if block.BlockHeader.Height <= maxPersistedHeight {
			logger.Debug().Msgf("Discarding block height %d, since node is ahead at height %d", block.BlockHeader.Height, maxPersistedHeight)
			continue
		}

		// Check if the block being synched is ahead of the current height
		if block.BlockHeader.Height > m.CurrentHeight() {
			logger.Info().Bool("TODO", true).Msgf("Received block at height %d, discarding as it is higher than the current height", block.BlockHeader.Height)
			// TECHDEBT: we need to store block responses that we are not yet ready to validate so we can validate them on a subsequent iteration of this loop
			// m.blocksResponsesReceived <- blockResponse
			continue
		}

		// Do basic block validation
		if err = m.validateBlock(block); err != nil {
			logger.Err(err).Msg("failed to validate block")
			continue
		}

		// Prepare the utility UOW of work to apply a new block
		if err := m.refreshUtilityUnitOfWork(); err != nil {
			m.logger.Error().Err(err).Msg("Could not refresh utility context")
			continue
		}

		// Update the leader proposing the block
		leaderIdInt, err := m.getNodeIdFromNodeAddress(string(block.BlockHeader.ProposerAddress))
		if err != nil {
			m.logger.Error().Err(err).Msg("Could not get leader id from leader address")
			continue
		}
		m.leaderId = typesCons.NewNodeId(leaderIdInt)

		// Try to apply the block by validating the transactions in the block
		if err := m.applyBlock(block); err != nil {
			m.logger.Error().Err(err).Msg("Could not apply block")
			continue
		}

		// Try to commit the block to persistence
		if err := m.commitBlock(block); err != nil {
			m.logger.Error().Err(err).Msg("Could not commit block")
			continue
		}
		logger.Info().Int64("height", int64(block.BlockHeader.Height)).Msgf("Block, at height %d is committed!", block.BlockHeader.Height)

		m.paceMaker.NewHeight()
		m.publishStateSyncBlockCommittedEvent(block.BlockHeader.Height)
	}
}

// metadataSyncLoop periodically sends metadata requests to its peers to aggregate metadata related to synching the state.
// It is intended to be run as a background process via `go metadataSyncLoop`
func (m *consensusModule) metadataSyncLoop() error {
	logger := m.logger.With().Str("source", "metadataSyncLoop").Logger()
	ctx := context.TODO()

	ticker := time.NewTicker(metadataSyncPeriod)
	for {
		select {
		case <-ticker.C:
			logger.Info().Msg("Background metadata sync check triggered")
			if err := m.broadcastMetadataRequests(); err != nil {
				logger.Error().Err(err).Msg("Failed to send metadata requests")
				return err
			}

		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

// broadcastMetadataRequests sends a metadata request to all peers in the network to understand
// the state of the network and determine if the node is behind.
func (m *consensusModule) broadcastMetadataRequests() error {
	stateSyncMetadataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}

	// TECHDEBT: This should be sent to all peers (full nodes, servicers, etc...), not just validators
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	for _, val := range validators {
		anyMsg, err := anypb.New(stateSyncMetadataReqMessage)
		if err != nil {
			return err
		}
		// TECHDEBT: Revisit why we're not using `Broadcast` here instead of `Send`.
		if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(val.GetAddress()), anyMsg); err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
			return err
		}
	}

	return nil
}

func (m *consensusModule) validateBlock(block *coreTypes.Block) error {
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

func (m *consensusModule) getAggregatedStateSyncMetadata() typesCons.StateSyncMetadataResponse {
	// TECHDEBT(#686): This should be an ongoing background passive state sync process but just
	// capturing the available messages at the time that this function was called is good enough for now.
	chanLen := len(m.metadataReceived)
	m.logger.Info().Msgf("Looping over %d state sync metadata responses", chanLen)

	minHeight, maxHeight := uint64(1), uint64(1)
	for i := 0; i < chanLen; i++ {
		metadata := <-m.metadataReceived
		if metadata.MaxHeight > maxHeight {
			maxHeight = metadata.MaxHeight
		}
		if metadata.MinHeight < minHeight {
			minHeight = metadata.MinHeight
		}
	}

	return typesCons.StateSyncMetadataResponse{
		PeerAddress: "unused_aggregated_metadata_address",
		MinHeight:   minHeight,
		MaxHeight:   maxHeight,
	}
}
