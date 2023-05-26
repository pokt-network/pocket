package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// StateSyncServerModule is responsible for handling requests and business logic that
// advertise and share local state metadata with other peers syncing to the latest block.
type StateSyncServerModule interface {
	// Advertise (send) the local state sync metadata to the requesting peer
	HandleStateSyncMetadataRequest(*typesCons.StateSyncMetadataRequest)

	// Advertise (send) the block being requested by the peer
	HandleGetBlockRequest(*typesCons.GetBlockRequest)
}

// HandleStateSyncMetadataRequest processes a request from another peer to get a view into the
// state stored in this node
func (m *stateSync) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) {
	logger := m.logger.With().Str("source", "HandleStateSyncMetadataRequest").Logger()

	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := metadataReq.PeerAddress

	// current height is the height of the block that is being processed, so we need to subtract 1 for the last finalized block
	prevPersistedBlockHeight := consensusMod.CurrentHeight() - 1

	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(prevPersistedBlockHeight))
	if err != nil {
		logger.Err(err).Msg("Error creating read context")
		return
	}
	defer readCtx.Release()

	// What is the maximum block height this node can share with others?
	maxHeight, err := readCtx.GetMaximumBlockHeight()
	if err != nil {
		logger.Err(err).Msg("Error getting max height")
		return
	}

	// What is the minimum block height this node can share with others?
	minHeight, err := readCtx.GetMinimumBlockHeight()
	if err != nil {
		logger.Err(err).Msg("Error getting min height")
		return
	}

	// Prepare state sync message to send to peer
	stateSyncMessage := typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataRes{
			MetadataRes: &typesCons.StateSyncMetadataResponse{
				PeerAddress: serverNodePeerAddress,
				MinHeight:   minHeight,
				MaxHeight:   maxHeight,
			},
		},
	}

	fields := map[string]interface{}{
		"max_height": maxHeight,
		"min_height": minHeight,
		"self":       serverNodePeerAddress,
		"peer":       clientPeerAddress,
	}

	if err = m.sendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress)); err != nil {
		logger.Err(err).Fields(fields).Msg("Error responding to state sync metadata request")
	}
	logger.Debug().Fields(fields).Msg("Successfully responded to state sync metadata request")
}

// HandleGetBlockRequest processes a request from another to share a specific block at a specific node
// that this node likely has available.
func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) {
	logger := m.logger.With().Str("source", "HandleGetBlockRequest").Logger()

	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerAddress := consensusMod.GetNodeAddress()
	clientPeerAddress := blockReq.PeerAddress

	// Check if the block should be retrievable based on the node's consensus height
	prevPersistedBlockHeight := consensusMod.CurrentHeight() - 1
	if prevPersistedBlockHeight < blockReq.Height {
		logger.Error().Msgf("The requested block height (%d) is higher than current persisted block height (%d)", blockReq.Height, prevPersistedBlockHeight)
		return
	}

	// Try to get block from the block store
	blockStore := m.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockStore.GetBlock(blockReq.Height)
	if err != nil {
		logger.Error().Err(err).Msgf("failed to get block at height %d", blockReq.Height)
		return
	}

	// Prepare state sync message to send to peer
	stateSyncMessage := typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockRes{
			GetBlockRes: &typesCons.GetBlockResponse{
				PeerAddress: serverNodePeerAddress,
				Block:       block,
			},
		},
	}

	fields := map[string]interface{}{
		"height": blockReq.Height,
		"self":   serverNodePeerAddress,
		"peer":   clientPeerAddress,
	}

	if err = m.sendStateSyncMessage(&stateSyncMessage, cryptoPocket.AddressFromString(clientPeerAddress)); err != nil {
		logger.Err(err).Fields(fields).Msg("Error responding to state sync block request")
	}
	logger.Debug().Fields(fields).Msg("Successfully responded to state sync block request")
}
