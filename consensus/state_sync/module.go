package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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

	logPrefix  string
	serverMode bool
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

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	//	m.aggregatedSyncMetadata = &typesCons.StateSyncMetadataResponse{}

	//	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)

	return m, nil
}

func (m *stateSync) Start() error {
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

	return m.Stop()
}

/*
pocket.StateMachineTransitionEvent
*/

func (m *stateSync) Stop() error {

	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
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
