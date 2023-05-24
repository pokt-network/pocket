package state_sync

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// This module is responsible for handling requests and business logic that advertises and shares
// local state metadata with other peers syncing to the latest block.
type StateSyncServerModule interface {
	// Advertise (send) the local state sync metadata to the requesting peer
	HandleStateSyncMetadataRequest(*typesCons.StateSyncMetadataRequest) error

	// Advertise (send) the block being requested by the peer
	HandleGetBlockRequest(*typesCons.GetBlockRequest) error
}

func (m *stateSync) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := metadataReq.PeerAddress

	m.logger.Info().Fields(m.stateSyncLogHelper(clientPeerAddress)).Msgf("Received StateSyncMetadataRequest %s", metadataReq)

	// current height is the height of the block that is being processed, so we need to subtract 1 for the last finalized block
	prevPersistedBlockHeight := consensusMod.CurrentHeight() - 1

	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(prevPersistedBlockHeight))
	if err != nil {
		return nil
	}
	defer readCtx.Release()

	maxHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil {
		return err
	}

	minHeight, err := readCtx.GetMinimumBlockHeight()
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

	return m.sendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress))
}

func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := blockReq.PeerAddress

	m.logger.Info().Fields(m.stateSyncLogHelper(clientPeerAddress)).Msgf("Received StateSync GetBlockRequest")
	prevPersistedBlockHeight := consensusMod.CurrentHeight() - 1

	if prevPersistedBlockHeight < blockReq.Height {
		return fmt.Errorf("requested block height: %d is higher than current persisted block height: %d", blockReq.Height, prevPersistedBlockHeight)
	}

	// get block from the persistence module
	blockStore := m.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockStore.GetBlock(blockReq.Height)
	if err != nil {
		m.logger.Error().Err(err).Msgf("failed to get block at height %d", blockReq.Height)
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

	return m.sendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress))
}
