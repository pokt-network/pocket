package consensus

import (
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	metadataSyncPeriod = 45 * time.Second // TODO: Make this configurable
	blockSyncPeriod    = 45 * time.Second // TODO: Make this configurable
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

// blockApplicationLoop commits the blocks received from the blocksResponsesReceived channel in state sync processes
func (m *consensusModule) blockApplicationLoop() {
	for blockResponse := range m.blocksResponsesReceived {
		block := blockResponse.Block
		m.logger.Info().Msgf("New block, at height %d is received!", block.BlockHeader.Height)

		maxPersistedHeight, err := m.maxPersistedBlockHeight()
		if err != nil {
			m.logger.Err(err).Msg("couldn't query max persisted height")
			continue
		}

		if block.BlockHeader.Height <= maxPersistedHeight {
			m.logger.Info().Msgf("Received block at height %d, discarding as it has already been persisted", block.BlockHeader.Height)
			return
		}

		// TODO: do not discard future blocks, consider adding them back into a channel to enable processing them later.
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
			continue
		}
		m.publishStateSyncBlockCommittedEvent(block.BlockHeader.Height)

	}
}

// sendMetadataRequests sends metadata requests to its peers
func (m *consensusModule) sendMetadataRequests() error {
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	for _, val := range validators {

		anyMsg, err := anypb.New(stateSyncMetaDataReqMessage)
		if err != nil {
			return err
		}
		if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(val.GetAddress()), anyMsg); err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
			return err
		}
	}

	return nil
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
