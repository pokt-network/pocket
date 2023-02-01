package state_sync

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
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
	serverNodePeerId, err := m.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	if err != nil {
		return err
	}

	clientPeerId := metadataReq.PeerId
	m.nodeLog(fmt.Sprintf("%s received State Sync MetaData Req from: %s", serverNodePeerId, clientPeerId))

	minHeight, err := m.GetBus().GetPersistenceModule().GetMinBlockHeight()
	if err != nil {
		return err
	}

	maxHeight, err := m.GetBus().GetPersistenceModule().GetMaxBlockHeight()
	if err != nil {
		return err
	}

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

	m.nodeLog(typesCons.SendingStateSyncMessage(&stateSyncMessage, clientPeerId, m.bus.GetConsensusModule().CurrentHeight()))
	return m.sendToPeer(anyMsg, cryptoPocket.AddressFromString(clientPeerId))
}

func (m *stateSync) HandleGetBlockRequest(blockReq *typesCons.GetBlockRequest) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerId, err := consensusMod.GetCurrentNodeAddressFromNodeId()
	if err != nil {
		return err
	}

	clientPeerId := blockReq.PeerId
	m.nodeLog(fmt.Sprintf("%s received State Sync Get Block Req from: %s", serverNodePeerId, clientPeerId))

	// IMPROVE: Consider checking the highest block from persistence, rather than the consensus module
	// check the max block height, if higher height is requested, return error
	if consensusMod.CurrentHeight() < blockReq.Height {
		return fmt.Errorf("requested block height: %d is higher than node's block height: %d", blockReq.Height, consensusMod.CurrentHeight())
	}

	// get block from the persistence module
	block, err := m.getBlockAtHeight(blockReq.Height)
	if err != nil {
		return err
	}

	stateSyncMessage := typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_RESPONSE,
		Message: &typesCons.StateSyncMessage_GetBlockRes{
			GetBlockRes: &typesCons.GetBlockResponse{
				PeerId: serverNodePeerId,
				Block:  block,
			},
		},
	}

	anyMsg, err := anypb.New(&stateSyncMessage)
	if err != nil {
		return err
	}
	m.nodeLog(typesCons.SendingStateSyncMessage(&stateSyncMessage, clientPeerId, blockReq.Height))
	return m.sendToPeer(anyMsg, cryptoPocket.AddressFromString(clientPeerId))
}

// Get a block from persistance module given block height
func (m *stateSync) getBlockAtHeight(blockHeight uint64) (*coreTypes.Block, error) {
	blockStore := m.GetBus().GetPersistenceModule().GetBlockStore()
	heightBytes := converters.HeightToBytes(blockHeight)

	blockBytes, err := blockStore.Get(heightBytes)
	if err != nil {
		m.nodeLog("Couldn't retrieve the block")
		return nil, err
	}

	var block coreTypes.Block
	err = codec.GetCodec().Unmarshal(blockBytes, &block)
	if err != nil {
		return &coreTypes.Block{}, err
	}

	return &block, nil
}
