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
)

// TODO_IN_THIS_COMMIT(goku): Move this into a README and add a diagram
// State sync implements synchronization for hotpokt blocks.
//
// Pocket node take one or two rules during state snychronization: client and/or server.
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
	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, dst cryptoPocket.Address, height uint64) error

	// Getter functions for the aggregated metadata and the metadata buffer, used by consensus module.
	GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse

	// Starts synching the node with the network by requesting blocks.
	TriggerSync() error
	PersistedBlock(uint64)

	// Returns the current state of the sync node.
	CurrentState() state

	HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error

	RequestMetadata() error
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

	if m.snycing { // if the node is currently syncing, update the sync state
		m.logger.Info().Msg("Node is already syncing, so updating the sync state.")
		m.state.endingHeight = m.aggregatedSyncMetadata.MaxHeight
	} else { // if the node is not currently syncing, generate a new sync state
		m.logger.Info().Msg("Node is currently not syncing, so generating a new sync state.")
		maxPersistedBlockHeight, err := m.maximumPersistedBlockHeight()
		if err != nil {
			return err
		}

		if maxPersistedBlockHeight > m.aggregatedSyncMetadata.MaxHeight || m.aggregatedSyncMetadata.MaxHeight == 0 {
			// should only happen when node is back online, or bootstraps, and the aggregated metadata is not updated yet.
			m.logger.Info().Msgf("NodeId: %d, Unsynched event is triggered, but aggregated metadata's height: %d is less than node's maxpersisted height: %d. So skipping the syncing. Syncing will start when there is a new block proposal and aggregated metadata is updated.", m.bus.GetConsensusModule().GetNodeId(), m.aggregatedSyncMetadata.MaxHeight, maxPersistedBlockHeight)
			return nil
		} else if maxPersistedBlockHeight == m.aggregatedSyncMetadata.MaxHeight {
			m.logger.Info().Msg("Node is already synched with the network, so skipping the syncing.")
			return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
		}

		m.snycing = true
		m.state = state{
			height:         maxPersistedBlockHeight,
			startingHeight: maxPersistedBlockHeight + 1,
			endingHeight:   m.aggregatedSyncMetadata.MaxHeight,
			blockReceived:  make(chan uint64, 1),
		}
		m.logger.Info().Msgf("Starting syncing from height %d to height %d", m.state.startingHeight, m.state.endingHeight)
		go m.Sync()
	}

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
	m := &stateSync{}

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

func (m *stateSync) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	m.logger.Info().Fields(m.logHelper(metaDataRes.PeerAddress)).Msgf("Received StateSync MetadataResponse: %s", metaDataRes)

	m.syncMetadataBuffer = append(m.syncMetadataBuffer, metaDataRes)
	//m.logger.Info().Msg("Finished handling StateSync MetadataResponse")

	return nil
}

func (m *stateSync) GetAggregatedStateSyncMetadata() *typesCons.StateSyncMetadataResponse {
	m.aggregatedSyncMetadata = m.aggregateMetadataResponses()
	return m.aggregatedSyncMetadata
}

func (m *stateSync) SetAggregatedSyncMetadata(metadata *typesCons.StateSyncMetadataResponse) {
	m.aggregatedSyncMetadata = metadata
}

// Snyc requests missing blocks one by one from its peers, and updates the syncing state (startingHeight, height, endingHeight).
// Sync listens on blockReceived channel, which sends heights of the persisted blocks received from peers. It uses this channel to update the height of the state: m.state.height = persistedBlockHeight
// if the received block is the target height, it will perform FSM state transition.
// else it will request the next block (after waiting sometime) and repeat the process.
func (m *stateSync) Sync() {
	m.logger.Info().Msg("Node is starting snycing...")

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

// Returns max block height metadainfo received from all peers by aggregating responses in the buffer.
func (m *stateSync) aggregateMetadataResponses() *typesCons.StateSyncMetadataResponse {
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

	//clear buffer
	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)

	return metadataResponse
}

// metadataSyncLoop periodically queries the network by sending metadata requests to peers using broadCastStateSyncMessage.
// CONSIDER: Improving meta data request synchronistaion, without timers.
func (m *stateSync) metadataSyncLoop() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.logger.Info().Msg("Periodic metadata sync is triggered")
			m.RequestMetadata()

		case <-ctx.Done():
			return nil

		}
	}
}

func (m *stateSync) RequestMetadata() error {
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}

	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	err := m.broadcastStateSyncMessage(stateSyncMetaDataReqMessage, currentHeight)
	if err != nil {
		return err
	}

	return nil
}
