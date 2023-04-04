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
	stateSyncModuleName = "stateSyncModule"
	metadataSyncPeriod  = 30 * time.Second // TODO: Make this configurable
)

// TODO_IN_THIS_COMMIT(goku): Move this into a README and/or add a diagram and/or improve readability.

// State sync implements synchronization for hotpokt blocks.
//
// Pocket node take one or two rules during state synchronization: client and/or server.
//
// There are two main processes run for the client role:
// 1. Metadata Aggregation loop (metadataSyncLoop)
//    - performs periodic metadata aggregation.
// 2. Requesting missing blocks (StartSyncing)
//    - snyncs by requesting missing blocks from peers.
// 3. Applying incoming blocks (HandleGetBlockResponse)
//    - applies missing blocks.

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule
	DebugStateSync

	// This functions are used for managing the Server mode of the node, which is handled independently from the FSM.
	IsServerModeEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	// TODO_IN_THIS_COMMIT: Do the functions below this line need to be part of the interface??/
	SendStateSyncMessage(msg *typesCons.StateSyncMessage, dst cryptoPocket.Address, height uint64) error

	// Getter functions for the aggregated metadata and the metadata buffer, used by consensus module.
	GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse

	// Starts synching the node with the network by requesting blocks.
	TriggerSync() error // TODO_IN_THIS_COMMIT: This should be an ongoing background process
	PersistedBlock(uint64)

	// Returns the current state of the sync node.
	CurrentState() state

	HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error
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

	serverModeEnabled bool

	// metadata buffer that is periodically updated
	aggregatedSyncMetadata *typesCons.StateSyncMetadataResponse // TECHDEBT: This needs a different type
	syncMetadataBuffer     []*typesCons.StateSyncMetadataResponse

	// Synching State
	state     state // TODO_IN_THIS_COMMIT: This is a property of the persistence module, it should not be part of consensus
	isSyncing bool  // TODO_IN_THIS_COMMIT: The FSM should be the source of truth for this

	// synchronization lifecycle controls
	ctx    context.Context
	cancel context.CancelFunc
}

// state represents the current "state" of the syncing.
type state struct {
	height         uint64 // latest persisted height, updated after every block is added to the module
	startingHeight uint64 // starting height for the sync, set when state is generated
	endingHeight   uint64 // ending height for the sync, set when state is generated

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
	m.logger.Info().Msg("Triggering syncing...")

	// if the node is currently syncing, update the sync state
	if m.isSyncing {
		m.logger.Info().Msg("Node is already syncing, so updating the sync state.")
		// 1. This should always be done automatically
		// 2. This doesn't trigger aggregation so it could be wrong
		m.state.endingHeight = m.aggregatedSyncMetadata.MaxHeight
		return nil
	}

	// if the node is not currently syncing, generate a new sync state
	m.logger.Info().Msg("Node is currently not syncing, so generating a new sync state.")
	maxPersistedBlockHeight, err := m.maximumPersistedBlockHeight()
	if err != nil {
		return err
	}

	if maxPersistedBlockHeight > m.aggregatedSyncMetadata.MaxHeight || m.aggregatedSyncMetadata.MaxHeight == 0 {
		// should only happen when node is back online, or bootstraps, and the aggregated metadata is not updated yet.
		// TODO_IN_THIS_COMMIT: This should be a warning, and make messages shorter; ditto everywhere else in this commit.
		m.logger.Info().Uint64("node_id", m.bus.GetConsensusModule().GetNodeId()).Msgf("Synched event is triggered, but aggregated metadata's height: %d is less than node's maxpersisted height: %d. So skipping the syncing. Syncing will start when there is a new block proposal and aggregated metadata is updated.", m.aggregatedSyncMetadata.MaxHeight, maxPersistedBlockHeight)
		return nil
	} else if maxPersistedBlockHeight == m.aggregatedSyncMetadata.MaxHeight {
		m.logger.Info().Msg("Node is already synched with the network, so skipping the syncing.")
		return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
	}

	m.isSyncing = true
	m.state = state{
		height:         maxPersistedBlockHeight,
		startingHeight: maxPersistedBlockHeight + 1,
		endingHeight:   m.aggregatedSyncMetadata.MaxHeight,
		blockReceived:  make(chan uint64, 1),
	}
	m.logger.Info().Msgf("Starting syncing from height %d to height %d", m.state.startingHeight, m.state.endingHeight)
	go m.Sync()

	return nil
}

// PersistedBlock is called by the consensus module's state_sync handler when a new block is received and persisted.
// It is used to update current height of the state that is being actively synched.
func (m *stateSync) PersistedBlock(blockHeight uint64) {
	m.logger.Info().Msgf("Block at height %d is persisted...", blockHeight)
	m.state.blockReceived <- blockHeight
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{
		serverModeEnabled:      false,
		aggregatedSyncMetadata: &typesCons.StateSyncMetadataResponse{},
		syncMetadataBuffer:     make([]*typesCons.StateSyncMetadataResponse, 0),
	}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	return m, nil
}

func (m *stateSync) Start() error {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	// Background process to periodically keep in sync with the state of the network
	go m.metadataSyncLoop()

	return nil
}

func (m *stateSync) Stop() error {
	m.cancel()
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

func (m *stateSync) IsServerModeEnabled() bool {
	return m.serverModeEnabled
}

func (m *stateSync) EnableServerMode() error {
	m.serverModeEnabled = true
	return nil
}

func (m *stateSync) DisableServerMode() error {
	m.serverModeEnabled = false
	return nil
}

func (m *stateSync) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	m.logger.Info().Fields(m.logHelper(metaDataRes.PeerAddress)).Msgf("Received StateSync MetadataResponse: %s", metaDataRes)

	m.syncMetadataBuffer = append(m.syncMetadataBuffer, metaDataRes)

	return nil
}

// TODO_IN_THIS_COMMIT: Does this need to be exposed?
func (m *stateSync) GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse {
	return m.aggregateMetadataResponses() // TODO_IN_THIS_COMMIT: Aggregation should happen in the background - it can just be done every time a new message comes in or using a channel; many alternatives available
}

// TODO_IN_THIS_COMMIT: Remove if we don't need this
func (m *stateSync) SetAggregatedSyncMetadata(metadata *typesCons.StateSyncMetadataResponse) {
	m.aggregatedSyncMetadata = metadata
}

// Sync requests missing blocks one by one from its peers, and updates the syncing state (startingHeight, height, endingHeight).
// Sync listens on blockReceived channel, which sends heights of the persisted blocks received from peers. It uses this channel to update the height of the state: m.state.height = persistedBlockHeight
// if the received block is the target height, it will perform FSM state transition.
// else it will request the next block (after waiting sometime) and repeat the process.
func (m *stateSync) Sync() {
	m.logger.Info().Msg("Node is starting syncing...")

	// Request blocks from the starting height to the ending height
	m.requestBlocks()

	// start timer to define a timeout period to re-request missing blocks.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		// check if the node is synched with the network
		case <-ticker.C:
			if m.state.height == m.state.endingHeight {
				m.logger.Info().Msgf("Node is synched for state: height: %d, starting height: %d, ending height: %d", m.state.height, m.state.startingHeight, m.state.endingHeight)
				break loop
			}
			// if not synched, request blocks again
			m.logger.Info().Msgf("Node is NOT synched for state: height: %d, starting height: %d, ending height: %d, requesting blocks", m.state.height, m.state.startingHeight, m.state.endingHeight)
			m.requestBlocks()

		// update the state with the received block height
		case persistedBlockHeight := <-m.state.blockReceived:
			m.state.height = persistedBlockHeight
			m.logger.Info().Msgf("Received block: %d, updating the state", persistedBlockHeight)

		case <-m.ctx.Done():
			return
		}
	}

	// TODO: this is temporary, check if node is validator, and transition accordingly
	m.logger.Info().Msg("Node is synched, transitions as validator \n")
	if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator); err != nil {
		m.logger.Err(err).Msg("Couldn't send state transition event")
	}
}

// requestBlocks requests missing blocks one by one from its peers.
func (m *stateSync) requestBlocks() {

	m.logger.Info().Msg("Requesting blocks")
	consensusMod := m.GetBus().GetConsensusModule()
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

}

// Returns max block height metadata info received from all peers by aggregating responses in the buffer.
func (m *stateSync) aggregateMetadataResponses() *typesCons.StateSyncMetadataResponse {
	aggregatedResponses := &typesCons.StateSyncMetadataResponse{}

	//aggregate metadataResponses by setting the metadataResponse
	// DISCUSS_IN_THIS_COMMIT: What if new responses come while we are processing this?
	// TODO_IN_THIS_OR_NEXT_COMMIT: `syncMetadataBuffer` needs to be a channel
	for _, meta := range m.syncMetadataBuffer {
		if meta.MaxHeight > aggregatedResponses.MaxHeight {
			aggregatedResponses.MaxHeight = meta.MaxHeight
		}
		if meta.MinHeight < aggregatedResponses.MinHeight {
			aggregatedResponses.MinHeight = meta.MinHeight
		}
	}

	m.logger.Debug().Uint64("min_height", aggregatedResponses.MinHeight).Uint64("max_height", aggregatedResponses.MaxHeight).Msg("Finished aggregated state sync metadata responses")

	// Clear buffer
	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)
	// Store aggregated responses
	m.aggregatedSyncMetadata = aggregatedResponses
	return m.aggregatedSyncMetadata
}

// metadataSyncLoop periodically queries the network to see if it is behind
func (m *stateSync) metadataSyncLoop() error {
	if m.ctx != nil {
		m.logger.Warn().Msg("metadataSyncLoop is already running. Cancelling the previous context...")
	}
	m.ctx, m.cancel = context.WithCancel(context.TODO())

	ticker := time.NewTicker(metadataSyncPeriod)
	for {
		select {
		case <-ticker.C:
			m.logger.Info().Msg("Background metadata sync check triggered")
			m.requestMetadata() // request more metadata from peers

		case <-m.ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (m *stateSync) requestMetadata() error {
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}

	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	return m.broadcastStateSyncMessage(stateSyncMetaDataReqMessage, currentHeight)
}
