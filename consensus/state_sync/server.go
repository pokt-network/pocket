package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"google.golang.org/protobuf/types/known/anypb"
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

// ! TODO implement
func (m *stateSyncModule) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) error {

	peerId := m.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	minHeight, maxHeight := m.aggregatedMetaResults()

	metadataRes := typesCons.StateSyncMetadataResponse{
		PeerId:    peerId,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}

	anyStateSyncMessage, err := anypb.New(&metadataRes)
	if err != nil {
		return err
	}

	return m.sendToPeer(anyStateSyncMessage, metadataReq.PeerId)
}

// ! TODO implement
func (m *stateSyncModule) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {

	// peerId := m.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()

	// metadataRes := typesCons.GetBlockResponse{
	// 	PeerId: peerId,
	// }

	return nil

}

// TODO! Placehalder functions for metadata aggregation of data received from different peers
func (m *stateSyncModule) aggregatedMetaResults() (uint64, uint64) {
	minHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	maxHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	return minHeight, maxHeight
}
