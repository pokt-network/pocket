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

	//! Placeholder implementation to show data types that can be used and flow, it will be replaced
	metadataRes := typesCons.StateSyncMetadataResponse{
		PeerId:    "1", // The `peer_id` needs to be populated by the P2P module of the receiving node so the sender cannot falsify its identity
		MinHeight: 0,
		MaxHeight: 10,
	}

	//m.sendToNode(metadataReq)

	anyStateSyncMessage, err := anypb.New(&metadataRes)
	if err != nil {
		return err
	}

	return m.sendToNode(anyStateSyncMessage)

}

// ! TODO implement, similar to sendToNode function of consensus. Maybe it can be used directly.
func (m *stateSyncModule) sendToNode(msg *anypb.Any) error {

	return nil
}

// ! TODO implement
func (m *stateSyncModule) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {
	return nil
}
