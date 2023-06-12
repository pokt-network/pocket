package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"google.golang.org/protobuf/proto"
)

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
	logger.Info().Int64("height", int64(block.BlockHeader.Height)).Msgf("Block, at height %d is committed!", block.BlockHeader.Height)

	m.publishStateSyncBlockCommittedEvent(block.BlockHeader.Height)
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
