package state_sync

import (
	"context"
	"sync"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	DefaultLogPrefix    = "NODE"
	stateSyncModuleName = "stateSyncModule"
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule
	DebugStateSync

	// This functions are used for managing the Server mode of the node, which is handled independently from the FSM.
	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	// This functions contains the business logic on handling Block and Metadata responses.
	//HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error
	//HandleGetBlockResponse(*typesCons.GetBlockResponse) error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, nodeAddress cryptoPocket.Address, height uint64) error

	SetStateSyncMetadataBuffer([]*typesCons.StateSyncMetadataResponse)
	GetStateSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse

	// Getter functions for the aggregated metadata and the metadata buffer, used by consensus module.
	GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse

	// Starts synching the node with the network by requesting blocks.
	StartSynching() error
}

// This interface should be only used for debugging purposes and tests.
type DebugStateSync interface {
	SetAggregatedSyncMetadata(*typesCons.StateSyncMetadataResponse)
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
	_ DebugStateSync        = &stateSync{}
)

type stateSync struct {
	bus    modules.Bus
	logger *modules.Logger

	m sync.RWMutex

	logPrefix  string
	serverMode bool

	// metadata buffer that is periodically updated
	aggregatedSyncMetadata *typesCons.StateSyncMetadataResponse
	syncMetadataBuffer     []*typesCons.StateSyncMetadataResponse
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{
		logPrefix: DefaultLogPrefix,
	}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.serverMode = false

	m.aggregatedSyncMetadata = &typesCons.StateSyncMetadataResponse{}

	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)

	return m, nil
}

func (m *stateSync) Start() error {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	// Node periodically checks if its up to date by requesting metadata from its peers as an external process with periodicMetadataSynch() function
	go m.periodicMetadataSync()

	return nil
}

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

func (m *stateSync) GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse {
	m.aggregatedSyncMetadata = m.aggregateMetadataResponses()
	return m.aggregatedSyncMetadata
}

func (m *stateSync) SetAggregatedSyncMetadata(metadata *typesCons.StateSyncMetadataResponse) {
	m.aggregatedSyncMetadata = metadata
}

func (m *stateSync) SetStateSyncMetadataBuffer(aggregatedSyncMetadata []*typesCons.StateSyncMetadataResponse) {
	m.m.Lock()
	defer m.m.Unlock()
	m.syncMetadataBuffer = aggregatedSyncMetadata
}

func (m *stateSync) GetStateSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse {
	return m.syncMetadataBuffer
}

// TODO(#352): Implement this function, currently a placeholder.
// Requests blocks one by one from its peers.
func (m *stateSync) StartSynching() error {
	consensusMod := m.GetBus().GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()

	currentPersistedBlockHeight, err := m.maximumPersistedBlockHeight()
	if err != nil {
		return err
	}

	maxPersistedBlockHeight := currentPersistedBlockHeight

	m.logger.Debug().Msgf("Starting synching, last persisted block height is: %d, aggregated metadata maxHeight is: %d", maxPersistedBlockHeight, m.aggregatedSyncMetadata.MaxHeight)

	// Request blocks one by one from its peers,
	for i := currentPersistedBlockHeight; i < m.aggregatedSyncMetadata.MaxHeight; i++ {
		blockHeightToRequest := i + 1

		// if blockHeightToRequest <= 2 {
		// 	blockHeightToRequest = 2
		// }

		m.logger.Debug().Msgf("StartSynching() Requesting block %d", blockHeightToRequest)
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      blockHeightToRequest,
				},
			},
		}
		m.broadcastStateSyncMessage(stateSyncGetBlockMessage, currentHeight)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// querying persistence to check if the requested block is received and persisted
		// loop:
		// 	for range ticker.C {
		// 		m.logger.Debug().Msg("Checking if block is persisted...")
		// 		maxPersistedBlockHeight, err = m.maximumPersistedBlockHeight()
		// 		if err != nil {
		// 			return err
		// 		}
		// 		if maxPersistedBlockHeight == blockHeightToRequest {
		// 			m.logger.Debug().Msgf("Block %d is persisted!", blockHeightToRequest)
		// 			break loop
		// 		}
		// 	}
	}

	isValidator, err := m.GetBus().GetConsensusModule().IsValidator()
	if err != nil {
		return err
	}

	if isValidator {
		return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSynchedValidator)
	}

	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSynchedNonValidator)
}

// TODO(#352): Implement this function, currently a placeholder.
// Returns max block height metadainfo received from all peers by aggregating responses in the buffer.
func (m *stateSync) aggregateMetadataResponses() *typesCons.StateSyncMetadataResponse {
	m.m.Lock()
	defer m.m.Unlock()

	metadataResponse := m.aggregatedSyncMetadata

	//aggregate metadataResponses by setting the metadataResponse
	for _, meta := range m.syncMetadataBuffer {
		if meta.MaxHeight > metadataResponse.MaxHeight {
			metadataResponse.MaxHeight = meta.MaxHeight
		}

		if meta.MinHeight < metadataResponse.MinHeight {
			metadataResponse.MinHeight = meta.MinHeight
		}
	}

	m.logger.Debug().Msgf("aggregateMetadataResponses, max height: %d", metadataResponse.MaxHeight)

	//clear the buffer
	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)

	return metadataResponse
}

// TODO(#352): Implement this function, currently a placeholder.
// Periodically (initially by using timers) queries the network by sending metadata requests to peers using broadCastStateSyncMessage() function.
// Update frequency can be tuned accordingly to the state. Initially, it will have a static timer for periodic snych.
// CONSIDER: Improving meta data request synchronistaion, without timers.
func (m *stateSync) periodicMetadataSync() error {

	//add timer channel with context to cancel the timer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// form a metaData request
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.logger.Debug().Msg("Periodic metadata synch is triggered")
			currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()

			err := m.broadcastStateSyncMessage(stateSyncMetaDataReqMessage, currentHeight)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return nil

		}
	}

	//return nil
}
