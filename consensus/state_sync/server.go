package state_sync

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// This module is responsible for handling requests and business logic that advertises and shares
// local state metadata with other peers synching to the latest block.
type StateSyncServerModule interface {
	// Advertise (send) the local state sync metadata to the requesting peer
	HandleStateSyncMetadataRequest(*typesCons.StateSyncMetadataRequest) error

	// Send the block being requested by the peer
	HandleGetBlockRequest(*typesCons.GetBlockRequest) error
}

func (m *stateSync) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerId := consensusMod.GetNodeAddress()
	currentHeight := consensusMod.CurrentHeight()
	clientPeerAddress := metadataReq.PeerAddress

	m.logger.Info().Msgf("%s received state sync metadata request from: %s, at height: %d \n", serverNodePeerId, clientPeerAddress, currentHeight)

	// current height is the height of the block that is being processed, so we need to subtract 1 for the last finalized block
	lastPersistedBlockHeight := currentHeight - 1
	persistenceContext, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(lastPersistedBlockHeight) - 1)
	if err != nil {
		return nil
	}
	defer persistenceContext.Close()

	maxHeight, err := persistenceContext.GetMaximumBlockHeight()
	if err != nil {
		return err
	}

	minHeight, err := persistenceContext.GetMinimumBlockHeight()
	if err != nil {
		return err
	}

	stateSyncMessage := typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataRes{
			MetadataRes: &typesCons.StateSyncMetadataResponse{
				PeerAddress: serverNodePeerId,
				MinHeight:   minHeight,
				MaxHeight:   uint64(maxHeight),
			},
		},
	}

	return m.SendStateSyncMessage(&stateSyncMessage, typesCons.StateSyncGetMetadataRequest, cryptoPocket.AddressFromString(clientPeerAddress), m.bus.GetConsensusModule().CurrentHeight())
}

func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := blockReq.PeerAddress
	currentHeight := consensusMod.CurrentHeight()

	m.logger.Info().Msgf("%s received state sync get block request from: %s, at height: %d \n", serverNodePeerAddress, clientPeerAddress, currentHeight)

	// current height is the height of the block that is being processed, so we need to subtract 1 for the last finalized block
	lastPersistedBlockHeight := currentHeight - 1
	if lastPersistedBlockHeight < blockReq.Height {
		return fmt.Errorf("requested block height: %d is higher than current persisted block height: %d (which is current height -1)", blockReq.Height, lastPersistedBlockHeight)
	}

	// get block from the persistence module
	block, err := m.getBlockAtHeight(blockReq.Height)
	if err != nil {
		return err
	}

	stateSyncMessage := typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockRes{
			GetBlockRes: &typesCons.GetBlockResponse{
				PeerAddress: serverNodePeerAddress,
				Block:       block,
			},
		},
	}

	return m.SendStateSyncMessage(&stateSyncMessage, typesCons.StateSyncGetBlockRequest, cryptoPocket.AddressFromString(clientPeerAddress), blockReq.Height)
}

// Get a block from persistence module given block height
func (m *stateSync) getBlockAtHeight(blockHeight uint64) (*coreTypes.Block, error) {
	blockStore := m.GetBus().GetPersistenceModule().GetBlockStore()
	heightBytes := converters.HeightToBytes(blockHeight)

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
