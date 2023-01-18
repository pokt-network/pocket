package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// This module is responsible for handling requests and business logic that advertises and shares
// local state metadata with other peers synching to the latest block.
type StateSyncServerModule interface {
	modules.Module

	// Advertise (send) the local state sync metadata to the requesting peer
	HandleStateSyncMetadataRequest(*typesCons.StateSyncMetadataRequest) error

	// Send the block being requested by the peer
	HandleGetBlockRequest(*typesCons.GetBlockRequest) error
}

type StateSyncModule interface {
	modules.Module

	// Handle a metadata response from a peer so this node can update its local view of the state
	// sync metadata available from its peers
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error

	// Handle a block response from a peer so this node can update apply it to its local state
	// and catch up to the global world state
	HandleStateSyncBlockResponse(*typesCons.StateSyncMetadataResponse) error
}

// ~~~~~ TODO(#362): The interface below is only meant to be used as a guideline and not implemented explicitly. It will updated & removed over time. ~~~~~
type StateSyncModuleLEGACY interface {
	// -- Constructor Setter Functions --

	// `HandleStateSync` function:
	// - Create a Utility Context
	// - Block.ValidateBasic()
	// - Consensus Module Replica Path
	//   - Prepare Block:  utilityContext.SetProposalBlock(block)
	//   - Apply Block:    utilityContext.ApplyBlock(block)
	//   - Validate Block: utilityContext.AppHash == Block.AppHash
	//   - Store Block:    consensusModule.CommitBlock()
	HandleStateSyncMessage(msg BlockResponseMessage)

	// `GetPeerSyncMeta` function:
	// - Retrieve a list of active peers with their metadata (identified and retrieved through P2P's `Churn Management`)
	GetPeerMetadata(GetPeerSyncMeta func() (peers []PeerSyncMeta, err error))

	// `NetworkSend` function contract:
	// - sends data to an address via P2P network
	NetworkSend(NetworkSend func(data []byte, address cryptoPocket.Address) error)

	// -- Sync modes --

	// In the StateSync protocol, the Node fields valid BlockRequests from its peers to help them CatchUp to be Synced.
	// This sub-protocol is continuous throughout the lifecycle of StateSync.
	RunServerMode()

	// In SyncedMode, the Node is caught up to the latest block and is listening & waiting for the latest block to be passed
	// to maintain a synchronous state with the global SyncState.
	// - UpdatePeerMetadata from P2P module
	// - UpdateSyncState
	// - Rely on new blocks to be propagated via the P2P network after Validators reach consensus
	// - If `localSyncState.Height < globalNetworkSyncState.Height` -> RunSyncMode() // careful about race-conditions
	RunSyncedMode()

	// Runs sync mode 'service' that continuously runs while `localSyncState.Height < globalNetworkSyncState.Height`
	// - UpdatePeerMetadata from P2P module
	// - Retrieve missing blocks from peers
	// - Process retrieved blocks
	// - UpdateSyncState
	// - If `localSyncState.Height == globalNetworkSyncState.Height` -> RunSyncedMode()
	RunSyncMode()

	// Returns the `highest priority aka lowest height` missing block heights up to `max` heights
	GetMissingBlockHeights(state SyncState, max int) (blockHeights []int64, err error)

	// Random selection of eligilbe peers enables a fair distribution of blockRequests over time via law of large numbers
	// An eligible peer is when `PeerMeta.MinHeight <= blockHeight <= PeerMeta.MaxHeight`
	GetRandomEligiblePeersForHeight(blockHeight int64) (eligiblePeer PeerSyncMeta, err error)

	// Uses `NetworkSend` to send a `BlockRequestMessage` to a specific peer
	SendBlockRequest(peerId string) error

	// Uses 'NetworkSend' to send a `BlockResponseMessage` to a specific peer
	// This function is used in 'ServerMode()'
	HandleBlockRequest(message BlockRequestMessage) error

	// Uses `HandleBlock` to process retrieved blocks from peers
	// Must update sync state using `SetMissingBlockHeight`
	ProcessBlock(block *coreTypes.Block) error
}

type SyncState interface {
	// latest local height
	LatestHeight() int64
	// latest network height (from the aggregation of Peer Sync Meta)
	LatestNetworkHeight() int64
	// retrieve peer meta (actively updated through churn management)
	GetPeers() []PeerSyncMeta
	// returns ordered array of missing block heights
	GetMissingBlockHeights() []int64
}

type BlockRequestMessage interface {
	// the height the peer wants from the block store
	GetHeight() int64
}

type BlockResponseMessage interface {
	// the bytes of the requested block from the block store
	GetBlockBytes() []byte
}

// TODO: needs to be shared between P2P as the Churn Management Process updates this information
type PeerSyncMeta interface {
	// the unique identifier associated with the peer
	GetPeerID() string
	// the maximum height the peer has in the block store
	GetMaxHeight() int64
	// the minimum height the peer has in the block store
	GetMinHeight() int64
}
