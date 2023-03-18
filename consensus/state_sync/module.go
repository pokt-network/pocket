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

// State sync implements synchronization for hotpokt blocks.
//
// Pocket node take one or two rules during state snychronization: client and/or server.
//
//
// There are three main processes run for the client role:
// 1. Metadata Aggregation loop (metadataSyncLoop)
//    - performs periodic metadata aggregation.
// 2. Requesting missing blocks (StartSynching)
//    - snyncs by requesting missing blocks from peers.
// 3. Applying incoming blocks (HandleGetBlockResponse)
//    - applies missing blocks.

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule
	DebugStateSync

	// This functions are used for managing the Server mode of the node, which is handled independently from the FSM.
	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, receiverPeerAddress cryptoPocket.Address, block_height uint64) error

	SetStateSyncMetadataBuffer([]*typesCons.StateSyncMetadataResponse)
	GetStateSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse

	// Getter functions for the aggregated metadata and the metadata buffer, used by consensus module.
	GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse

	// Starts synching the node with the network by requesting blocks.
	TriggerSync() error
	PersistedBlock(uint64)

	// Returns the current state of the sync node.
	CurrentState() state

	// Set the current state of the node.
	//SetStateCurrentHeight(uint64)
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

	// current state that is synched, or previously last synched
	state   state
	snycing bool

	logPrefix  string
	serverMode bool

	// metadata buffer that is periodically updated
	aggregatedSyncMetadata *typesCons.StateSyncMetadataResponse
	syncMetadataBuffer     []*typesCons.StateSyncMetadataResponse

	// snychronisation lifecycle controls
	ctx    context.Context
	cancel context.CancelFunc

	// signal only channel triggered by the FSM transition
	triggerSyncChan chan struct{}
}

// information about current state of the synching node
type state struct {
	height         uint64 // latest persisted height, updated after every block is added to the module
	startingHeight uint64 // starting height for the sync, set when state is generated
	endingHeight   uint64 // ending height for the sync, set when state is generated
	err            error  //snyc error

	blockReceived chan uint64
}

func (m *stateSync) CurrentState() state {
	return m.state
}

// TriggerSync is the entry point of state sync module for synching. It performs two tasks:
//
// 1. If the node is not currently syncing, it generates a new sync state. And triggers the sync.
// 2. If the node is currently syncing, it updates the current sync state, by updating the ending height.
func (m *stateSync) TriggerSync() error {
	// check if the node is not currently syncing, if it is synching update the state
	m.m.Lock()
	if m.snycing { // if the node is currently syncing, update the sync state
		m.state.endingHeight = m.aggregatedSyncMetadata.MaxHeight
	} else { // if the node is not currently syncing, generate a new sync state
		maxPersistedBlockHeight, err := m.maximumPersistedBlockHeight()
		if err != nil {
			return err
		}
		m.snycing = true
		m.state = state{
			height:         maxPersistedBlockHeight,
			startingHeight: maxPersistedBlockHeight,
			endingHeight:   m.aggregatedSyncMetadata.MaxHeight,
			blockReceived:  make(chan uint64, 1),
		}
	}
	m.m.Unlock()

	// signal the sync loop to start
	select {
	case m.triggerSyncChan <- struct{}{}:
	default:
	}

	return nil
}

// PersistedBlock is called by the consensus module's state_sync handler when a new block is received and persisted.
// It is used to update current height of the state that is being actively synched.
func (m *stateSync) PersistedBlock(blockHeight uint64) {
	m.state.blockReceived <- blockHeight
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

	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Node periodically checks if its up to date by requesting metadata from its peers as an external process with periodicMetadataSynch() function
	go m.metadataSyncLoop()

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

// StartSynching
// requests missing blocks one by one from its peers.
// makes checks on blockReceived channel to see which block is received and persisted.
// with each reiceved block it will update the current height of the state that is being actively synched.
// if the received block is the target height, it will perform FSM state transition.
// else it will request the next block (after waiting sometime) and repeat the process.
// check how others handle re-requesting blocks
func (m *stateSync) Sync() error {

	consensusMod := m.GetBus().GetConsensusModule()
	nodeAddress := consensusMod.GetNodeAddress()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			m.m.Lock()
			if m.state.height == m.state.endingHeight {
				m.logger.Log().Msgf("Node is synched for state, height: %d, starting height: %d, ending height:%d", m.state.height, m.state.startingHeight, m.state.startingHeight)
				break loop
			}
			m.m.Unlock()

			// start requesting the missing blocks in the state
			// fire and forget pattern, broadcasts to all peers
			blockToRequest := m.state.height + 1
			for i := blockToRequest; i <= m.state.endingHeight; i++ {
				m.logger.Log().Msgf("Sync is requesting block: %d", i)
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

		case persistedBlockHeight := <-m.state.blockReceived:
			m.m.Lock()
			m.state.height = persistedBlockHeight
			m.m.Unlock()
			m.logger.Log().Msgf("Received block: %d, updating the state", persistedBlockHeight)

		case <-m.ctx.Done():
			return nil

		}
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

// Periodically (initially by using timers) queries the network by sending metadata requests to peers using broadCastStateSyncMessage() function.
// Update frequency can be tuned accordingly to the state. Initially, it will have a static timer for periodic snych.
// CONSIDER: Improving meta data request synchronistaion, without timers.
func (m *stateSync) metadataSyncLoop() error {

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

	ticker := time.NewTicker(10 * time.Second)
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
}

func (m state) isFinished() bool {
	return m.endingHeight <= m.height
}

// Consider SyncWait()
