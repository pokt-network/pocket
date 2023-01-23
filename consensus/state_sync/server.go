package state_sync

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
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

func (m *stateSync) HandleStateSyncMetadataRequest(metadataReq *typesCons.StateSyncMetadataRequest) error {

	serverNodePeerId, err := m.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	if err != nil {
		return err
	}

	clientPeerId := metadataReq.PeerId
	m.nodeLog(fmt.Sprintf("%s Received State Sync MetaData Req from: %s", serverNodePeerId, clientPeerId))

	minHeight, maxHeight := m.aggregateMetaResults()

	stateSyncMessage := typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE,
		Message: &typesCons.StateSyncMessage_MetadataRes{
			MetadataRes: &typesCons.StateSyncMetadataResponse{
				PeerId:    serverNodePeerId,
				MinHeight: minHeight,
				MaxHeight: maxHeight,
			},
		},
	}

	anyMsg, err := anypb.New(&stateSyncMessage)
	if err != nil {
		return err
	}

	return m.sendToPeer(anyMsg, cryptoPocket.AddressFromString(clientPeerId))
}

func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {
	serverNodePeerId, err := m.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	if err != nil {
		return err
	}

	clientPeerId := blockReq.PeerId

	m.nodeLog(fmt.Sprintf("%s Received State Sync Get Block Req from: %s", serverNodePeerId, clientPeerId))

	block, err := m.getBlockAtHeight(blockReq.Height)
	if err != nil {
		return err
	}

	metadataRes := typesCons.GetBlockResponse{
		PeerId: serverNodePeerId,
		Block:  block,
	}

	anyStateSyncMessage, err := anypb.New(&metadataRes)
	if err != nil {
		return err
	}

	return m.sendToPeer(anyStateSyncMessage, cryptoPocket.AddressFromString(blockReq.PeerId))
}

// TODO IMPLEMENT
// Placeholder function for metadata aggregation of data received from different peers
func (m *stateSync) aggregateMetaResults() (uint64, uint64) {
	minHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	maxHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	return minHeight, maxHeight
}

// Get a block from persistance module given block height
func (m *stateSync) getBlockAtHeight(blockHeight uint64) (*coreTypes.Block, error) {
	blockStore := m.GetBus().GetPersistenceModule().GetBlockStore()
	heightBytes := heightToBytes(int64(blockHeight))

	blockBytes, err := blockStore.Get(heightBytes)

	if err != nil {
		return &coreTypes.Block{}, err
	}

	var block coreTypes.Block
	err = codec.GetCodec().Unmarshal(blockBytes, &block)
	if err != nil {
		return &coreTypes.Block{}, err
	}

	return &block, nil
}
