package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
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
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := metadataReq.PeerAddress

	m.logger.Info().Fields(m.logHelper(clientPeerAddress)).Msgf("Received StateSyncMetadataRequest %s", metadataReq)

	maxHeight, err := m.maximumPersistedBlockHeight()
	if err != nil {
		return err
	}

	minHeight, err := m.minimumPersistedBlockHeight()
	if err != nil {
		return err
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

	return m.SendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress), m.bus.GetConsensusModule().CurrentHeight())
}

func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := blockReq.PeerAddress

	lastPersistedBlockHeight, err := m.maximumPersistedBlockHeight()
	if err != nil {
		return err
	}

	m.logger.Info().Fields(m.logHelper(clientPeerAddress)).Msgf("Received StateSync GetBlockRequest: %s", blockReq)

	if lastPersistedBlockHeight < blockReq.Height {
		m.logger.Debug().Msgf("requested block height: %d is higher than current persisted block height: %d, so not going to reply", blockReq.Height, lastPersistedBlockHeight)
		return nil
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

	//m.logger.Info().Fields(m.logHelper(clientPeerAddress)).Msgf("Responding to StateSync GetBlockRequest Height: %x,\n Transactions: %x", block.BlockHeader.Height, &stateSyncMessage.GetGetBlockRes().Block.Transactions)

	return m.SendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress), blockReq.Height)
}

// Get a block from persistence module given block height
func (m *stateSync) getBlockAtHeight(blockHeight uint64) (*coreTypes.Block, error) {
	//m.logger.Debug().Msgf("Trying to get block at height: %d", blockHeight)
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
