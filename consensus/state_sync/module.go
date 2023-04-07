package state_sync

import (
	"sync"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName = "stateSyncModule"
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule
	//DebugStateSync

	// This functions are used for managing the Server mode of the node, which is handled independently from the FSM.
	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	// This functions contains the business logic on handling Block and Metadata responses.
	//HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error
	//HandleGetBlockResponse(*typesCons.GetBlockResponse) error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, nodeAddress cryptoPocket.Address, height uint64) error

	// Getter functions for the aggregated metadata and the metadata buffer, used by consensus module.
	//GetAggregatedMetadata() *typesCons.StateSyncMetadataResponse
}

// // This interface should be only used for debugging purposes and tests.
// type DebugStateSync interface {
// 	SetAggregatedMetadata(*typesCons.StateSyncMetadataResponse)
// }

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
	//_ DebugStateSync        = &stateSync{}
)

type stateSync struct {
	bus    modules.Bus
	logger *modules.Logger

	m sync.RWMutex

	logPrefix  string
	serverMode bool

	//targetBlockHeight uint64

	// metadata buffer that is periodically updated
	//aggregatedSyncMetadata *typesCons.StateSyncMetadataResponse
	//syncMetadataBuffer     []*typesCons.StateSyncMetadataResponse

}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.serverMode = false

	//	m.aggregatedSyncMetadata = &typesCons.StateSyncMetadataResponse{}

	//	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)

	return m, nil
}

func (m *stateSync) Start() error {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())
	//get metadatas from metadata channel

	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	_, maxHeight := consensusMod.GetAggregatedStateSyncMetadata()
	nodeAddress := consensusMod.GetNodeAddress()

	// request blocks starting from currentHeight to maxHeight, wait for blockReceived channel for each block
	for i := currentHeight; i < maxHeight; i++ {
		m.logger.Info().Msgf("Requesting block %d", i)
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      i,
				},
			},
		}
		m.broadcastStateSyncMessage(stateSyncGetBlockMessage, i)

		// wait for 5 seconds with a timer to receive the block.

	}

	return nil
}

/*
// Start starts syncing the node with the network by requesting blocks.
func (m *stateSync) StartSyncing() error {
	//m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	//current_height := m.GetBus().GetConsensusModule().CurrentHeight()

	for i := currentHeight; i < targetBlockHeight; i++ {
		fmt.Println("Requesting block", i)

	}

	nodeAddress := consensusMod.GetNodeAddress()

	// start requesting the missing blocks in the state
	// fire and forget pattern, broadcasts to all peers
	blockToRequest := m.state.height + 1
	for i := blockToRequest; i <= m.state.endingHeight; i++ {
		m.logger.Info().Msgf("Sync is requesting block: %d, starting height: %d, ending height: %d", i, m.state.startingHeight, m.state.endingHeight)
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      i,
				},
			},
		}
		m.broadcastStateSyncMessage(stateSyncGetBlockMessage, i)
	}

	// Node periodically checks if its up to date by requesting metadata from its peers as an external process with periodicMetadataSync() function
	//go m.periodicMetadataSync()

	return nil
}
*/

func (m *stateSync) Stop() error {
	return nil
}

func (m *stateSync) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *stateSync) GetBus() modules.Bus {
	if m.bus == nil {
		logger.Global.Fatal().Msg("PocketBus is not initialized")
	}
	return m.bus
}

func (m *stateSync) GetModuleName() string {
	return stateSyncModuleName
}

func (m *stateSync) IsServerModEnabled() bool {
	return m.serverMode
}

func (m *stateSync) SetLogPrefix(logPrefix string) {
	m.logPrefix = logPrefix
}

func (m *stateSync) EnableServerMode() error {
	m.serverMode = true
	return nil
}

func (m *stateSync) DisableServerMode() error {
	m.serverMode = false
	return nil
}

// func (m *stateSync) GetAggregatedMetadata() *typesCons.StateSyncMetadataResponse {
// 	m.aggregatedSyncMetadata = m.aggregateMetadataResponses()
// 	return m.aggregatedSyncMetadata
// }

// func (m *stateSync) SetAggregatedMetadata(metadata *typesCons.StateSyncMetadataResponse) {
// 	m.aggregatedSyncMetadata = metadata
// }

// // TODO(#352): Implement the business logic for this function
// // Requests blocks one by one from its peers.
// func (m *stateSync) StartSyncing() error {
// 	current_height := m.GetBus().GetConsensusModule().CurrentHeight()

// 	m.logger.Debug().Msgf("Starting syncing, current height %d, aggregated maxHeight %d", current_height, m.aggregatedSyncMetadata.MaxHeight)

// 	// TODO: Implement the business logic for this function
// 	//mockSyncing()
// 	//m.GetBus().GetConsensusModule().SetHeight(m.aggregatedSyncMetadata.MaxHeight)
// 	//h, r, s := m.bus.GetConsensusModule().CurrentHeight(), m.bus.GetConsensusModule().CurrentRound(), uint8(m.bus.GetConsensusModule().CurrentStep())
// 	//fmt.Printf("h: %d, r: %d, s:  %d, leaderId is set: %d", h, r, s, m.bus.GetConsensusModule().GetLeaderForView(h, r, s))
// 	return nil
// }

// TODO(#352):  Implement the business logic for this function
// Returns max block height metadainfo received from all peers by aggregating responses in the buffer.
// func (m *stateSync) aggregateMetadataResponses() *typesCons.StateSyncMetadataResponse {
// 	m.m.Lock()
// 	defer m.m.Unlock()

// 	metadataResponse := m.aggregatedSyncMetadata

// 	return metadataResponse
// }

// TODO(#352): Implement this function, currently a placeholder.
// Periodically (initially by using timers) queries the network by sending metadata requests to peers using broadCastStateSyncMessage() function.
// func (m *stateSync) periodicMetadataSync() {

// 	// uses a timer to periodically query the network
// 	// form a metadata request
// 	// send to peers using broadCastStateSyncMessage()
// }
