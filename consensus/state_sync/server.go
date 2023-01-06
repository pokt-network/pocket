package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

// This module is responsible for handling requests and business logic that advertises and shares
// local state metadata with other peers synching to the latest block.
type StateSyncServerModule interface {
	//modules.Module

	// Advertise (send) the local state sync metadata to the requesting peer
	HandleStateSyncMetadataRequest(*typesCons.StateSyncMetadataRequest) error

	// Send the block being requested by the peer
	HandleGetBlockRequest(*typesCons.GetBlockRequest) error
}

func (m *stateSyncModule) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) error {
	return nil
}

func (m *stateSyncModule) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {

	return nil
}
