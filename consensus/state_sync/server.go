package state_sync

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
)

// This module is responsible for handling requests and business logic that advertises and shares
// local state metadata with other peers syncing to the latest block.
type StateSyncServerModule interface {
	// Advertise (send) the local state sync metadata to the requesting peer
	HandleStateSyncMetadataRequest(*typesCons.StateSyncMetadataRequest)

	// Advertise (send) the block being requested by the peer
	HandleGetBlockRequest(*typesCons.GetBlockRequest)
}

func (m *stateSync) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := metadataReq.PeerAddress

	m.logger.Info().Fields(m.stateSyncLogHelper(clientPeerAddress)).Msgf("Handling StateSync MetadataRequest")

	// current height is the height of the block that is being processed, so we need to subtract 1 for the last finalized block
	prevPersistedBlockHeight := consensusMod.CurrentHeight() - 1

	// TODO check if we need to send currentRound here rather than prevPersistedBlockHeight
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(prevPersistedBlockHeight))
	if err != nil {
		m.logger.Err(err).Msg("Error creating read context")
		return
	}
	defer readCtx.Release()

	maxHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil {
		m.logger.Err(err).Msg("Error getting max height")
		return
	}

	minHeight, err := readCtx.GetMinimumBlockHeight()
	if err != nil {
		m.logger.Err(err).Msg("Error getting min height")
		return
	}

	stateSyncMessage := typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataRes{
			MetadataRes: &typesCons.StateSyncMetadataResponse{
				PeerAddress: serverNodePeerAddress,
				MinHeight:   minHeight,
				MaxHeight:   uint64(maxHeight),
			},
		},
	}

	err = m.sendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress))
	if err != nil {
		m.logger.Err(err).Msg("Error sending state sync message")
		return
	}
}

func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := blockReq.PeerAddress

	m.logger.Info().Fields(m.stateSyncLogHelper(clientPeerAddress)).Msgf("Handling StateSync GetBlockRequest")
	prevPersistedBlockHeight := consensusMod.CurrentHeight() - 1

	if prevPersistedBlockHeight < blockReq.Height {
		m.logger.Err(fmt.Errorf("requested block height: %d is higher than current persisted block height: %d", blockReq.Height, prevPersistedBlockHeight))
		return
	}

	// get block from the persistence module
	block, err := m.getBlockAtHeight(blockReq.Height)
	if err != nil {
		m.logger.Err(err).Msg("Error getting block")
		return
	}

	stateSyncMessage := typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockRes{
			GetBlockRes: &typesCons.GetBlockResponse{
				PeerAddress: serverNodePeerAddress,
				Block:       block,
			},
		},
	}

	err = m.sendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress))
	if err != nil {
		m.logger.Err(err).Msg("Error sending state sync message")
		return
	}
}

// Get a block from persistence module given block height
func (m *stateSync) getBlockAtHeight(blockHeight uint64) (*coreTypes.Block, error) {
	blockStore := m.GetBus().GetPersistenceModule().GetBlockStore()
	heightBytes := utils.HeightToBytes(blockHeight)

	blockBytes, err := blockStore.Get(heightBytes)
	if err != nil {
		m.logger.Error().Err(typesCons.ErrConsensusMempoolFull).Msg(typesCons.DisregardHotstuffMessage)
		return nil, err
	}

	var block coreTypes.Block
	err = codec.GetCodec().Unmarshal(blockBytes, &block)
	if err != nil {
		return &coreTypes.Block{}, err
	}

	return &block, nil
}
