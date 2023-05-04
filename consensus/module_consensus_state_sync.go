package consensus

import (
	"context"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const metadataSyncPeriod = 45 * time.Second // TODO: Make this configurable

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

// blockApplicationLoop commits the blocks received from the blocksResponsesReceived channel
// it is intended to be run as a background process
func (m *consensusModule) blockApplicationLoop() {
	logger := m.logger.With().Str("source", "blockApplicationLoop").Logger()

	for blockResponse := range m.blocksResponsesReceived {
		block := blockResponse.Block
		logger.Info().Msgf("New block, at height %d is received!", block.BlockHeader.Height)

		maxPersistedHeight, err := m.maxPersistedBlockHeight()
		if err != nil {
			logger.Err(err).Msg("couldn't query max persisted height")
			continue
		}

		// CONSIDERATION: rather than discarding these blocks, push them into a channel to process them later
		if block.BlockHeader.Height <= maxPersistedHeight {
			logger.Info().Msgf("Received block at height %d, discarding as it has already been persisted", block.BlockHeader.Height)
			continue
		}

		if block.BlockHeader.Height > m.CurrentHeight() {
			logger.Info().Msgf("Received block at height %d, discarding as it is higher than the current height", block.BlockHeader.Height)
			continue
		}

		if err = m.validateBlock(block); err != nil {
			logger.Err(err).Msg("failed to validate block")
			continue
		}

		if err = m.applyAndCommitBlock(block); err != nil {
			logger.Err(err).Msg("failed to apply and commit block")
			continue
		}
		logger.Info().Msgf("Block, at height %d is committed!", block.BlockHeader.Height)
		m.publishStateSyncBlockCommittedEvent(block.BlockHeader.Height)
	}
}

// metadataSyncLoop periodically sends metadata requests to its peers
// it is intended to be run as a background process
func (m *consensusModule) metadataSyncLoop() error {
	ctx := context.TODO()

	ticker := time.NewTicker(metadataSyncPeriod)
	for {
		select {
		case <-ticker.C:
			m.logger.Info().Msg("Background metadata sync check triggered")
			if err := m.sendMetadataRequests(); err != nil {
				m.logger.Error().Err(err).Msg("Failed to send metadata requests")
				return err
			}

		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

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

	if err := m.refreshUtilityUnitOfWork(); err != nil {
		m.logger.Error().Err(err).Msg("Could not refresh utility context")
		return err
	}

	leaderIdInt, err := m.GetNodeIdFromNodeAddress(string(block.BlockHeader.ProposerAddress))
	if err != nil {
		m.logger.Error().Err(err).Msg("Could not get leader id from leader address")
		return err
	}

	leaderId := typesCons.NodeId(leaderIdInt)
	m.leaderId = &leaderId

	return nil
}

func (m *consensusModule) applyAndCommitBlock(block *coreTypes.Block) error {
	if err := m.applyBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not apply block, invalid QC")
		return err
	}

	if err := m.commitBlock(block); err != nil {
		m.logger.Error().Err(err).Msg("Could not commit block, invalid QC")
		return err
	}
	m.paceMaker.NewHeight()

	m.logger.Info().Msgf("New block is committed, current height is :%d", m.height)
	return nil
}

func (m *consensusModule) getAggregatedStateSyncMetadata() typesCons.StateSyncMetadataResponse {
	minHeight, maxHeight := uint64(1), uint64(1)
	chanLen := len(m.metadataReceived)

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
