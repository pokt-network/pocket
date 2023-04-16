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

const metadataSyncPeriod = 30 * time.Second // TODO: Make this configurable

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

// commitReceivedBlocks commits the blocks received from the blocksReceived channel
// it runs as a background process in consensus module
// listens on the blocksReceived channel, verifies and commits the received block
func (m *consensusModule) blockApplicationLoop() {
	for blockResponse := range m.blocksReceived {
		block := blockResponse.Block
		m.logger.Info().Msgf("New block, at height %d is received!", block.BlockHeader.Height)

		maxPersistedHeight, err := m.maxPersistedBlockHeight()
		if err != nil {
			m.logger.Err(err).Msg("couldn't query max persisted height")
			continue
		}

		//fmt.Println("Now going to decide if I should apply it")
		if block.BlockHeader.Height <= maxPersistedHeight {
			m.logger.Info().Msgf("Received block with height: %d, but node already persisted blocks until height: %d, so node will not apply this block", block.BlockHeader.Height, maxPersistedHeight)
			continue
		} else if block.BlockHeader.Height > m.CurrentHeight() {
			m.logger.Info().Msgf("Received block with height %d, but node's last persisted height is: %d, so node will not apply this block", block.BlockHeader.Height, maxPersistedHeight)
			continue
		}

		err = m.verifyBlock(block)
		if err != nil {
			m.logger.Err(err).Msg("failed to verify block")
			continue
		}

		err = m.applyAndCommitBlock(block)
		if err != nil {
			m.logger.Err(err).Msg("failed to apply and commit block")
			continue
		}
		m.logger.Info().Msgf("Block, at height %d is committed!", block.BlockHeader.Height)
		m.stateSync.CommittedBlock(m.CurrentHeight())
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
			m.sendMetadataRequests()

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

// TODO! If verify block tries to verify, state sync tests will fail as state sync blocks are empty.
func (m *consensusModule) verifyBlock(block *coreTypes.Block) error {
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

	maxPersistedHeight, err := m.maxPersistedBlockHeight()
	if err != nil {
		return err
	}

	m.logger.Info().Msgf("Block is Committed, maxPersistedHeight is: %d, current height is :%d", maxPersistedHeight, m.height)
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
