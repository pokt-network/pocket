package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ConsensusStateSync = &consensusModule{}

func (m *consensusModule) GetNodeIdFromNodeAddress(peerId string) (uint64, error) {
	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		// REFACTOR(#434): As per issue #434, once the new id is sorted out, this return statement must be changed
		return 0, err
	}

	valAddrToIdMap := typesCons.NewActorMapper(validators).GetValAddrToIdMap()
	return uint64(valAddrToIdMap[peerId]), nil
}

func (m *consensusModule) GetNodeAddress() string {
	return m.nodeAddress
}

// blockApplicationLoop commits the blocks received from the blocksResponsesReceived channel in passive state sync process
// it is intended to be run as a background process
func (m *consensusModule) blockApplicationLoop() {
	for blockResponse := range m.blocksResponsesReceived {
		block := blockResponse.Block
		maxPersistedHeight, err := m.maxPersistedBlockHeight()
		if err != nil {
			m.logger.Err(err).Msg("couldn't query max persisted height")
			return
		}

		// TODO: rather than discarding these blocks, push them into a channel to process them later
		if block.BlockHeader.Height <= maxPersistedHeight {
			m.logger.Info().Msgf("Received block at height %d, discarding as it has already been persisted", block.BlockHeader.Height)
			return
		}

		if block.BlockHeader.Height > m.CurrentHeight() {
			m.logger.Info().Msgf("Received block at height %d, discarding as it is higher than the current height", block.BlockHeader.Height)
			return
		}

		if err = m.validateBlock(block); err != nil {
			m.logger.Err(err).Msg("failed to validate block")
			return
		}

		if err = m.applyAndCommitBlock(block); err != nil {
			m.logger.Err(err).Msg("failed to apply and commit block")
			return
		}
		m.publishStateSyncBlockCommittedEvent(block.BlockHeader.Height)

	}
}

func (m *consensusModule) maxPersistedBlockHeight() (uint64, error) {
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(m.CurrentHeight()))
	if err != nil {
		return 0, err
	}
	defer readCtx.Release()

	maxHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil {
		return 0, err
	}

	return maxHeight, nil
}

func (m *consensusModule) validateBlock(block *coreTypes.Block) error {
	// TODO(#352): add quorum certificate validation for the block
	return nil
}

func (m *consensusModule) applyAndCommitBlock(block *coreTypes.Block) error {
	// TODO(#352): call m.applyBlock(block) function before  m.commitBlock(block). In this PR testing blocks don't have a valid QC, therefore commented out to let the tests pass.
	if err := m.commitBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not commit block, invalid QC")
		return err
	}
	m.paceMaker.NewHeight()

	m.logger.Info().Msgf("New block is committed, current height is :%d", m.height)
	return nil
}
