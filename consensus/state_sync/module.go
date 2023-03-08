package state_sync

import (
	"sync"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
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

	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error
	HandleGetBlockResponse(*typesCons.GetBlockResponse) error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, nodeAddress cryptoPocket.Address, height uint64) error

	GetSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse
	GetAggregatedSyncMetadata() *typesCons.StateSyncMetadataResponse

	StartSynching() error
}

// This interface is used for debugging purposes.
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

	// Node periodically checks if its up to date by requesting metadata from its peers as an external process with periodicMetaDataSynch() function
	go m.periodicMetaDataSynch()

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

func (m *stateSync) GetAggregatedSyncMetadata() *typesCons.StateSyncMetadataResponse {
	m.aggregatedSyncMetadata = m.aggregateMetadataResponses()
	return m.aggregatedSyncMetadata
}

func (m *stateSync) SetAggregatedSyncMetadata(metaData *typesCons.StateSyncMetadataResponse) {
	m.aggregatedSyncMetadata = metaData
}

// TODO(#352): Implement this function, currently a placeholder.
func (m *stateSync) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerId := consensusMod.GetNodeAddress()
	clientPeerId := metaDataRes.PeerAddress
	currentHeight := consensusMod.CurrentHeight()

	fields := map[string]any{
		"currentHeight": currentHeight,
		"sender":        serverNodePeerId,
		"receiver":      clientPeerId,
	}

	m.logger.Info().Fields(fields).Msgf("Received StateSyncMetadataResponse: %s", metaDataRes)

	m.m.Lock()
	defer m.m.Unlock()

	return nil
}

// TODO(#352): Implement this function, currently a placeholder.
func (m *stateSync) HandleGetBlockResponse(blockRes *typesCons.GetBlockResponse) error {

	serverNodePeerId := m.bus.GetConsensusModule().GetNodeAddress()
	clientPeerId := blockRes.PeerAddress

	fields := map[string]any{
		"currentHeight": blockRes.Block.BlockHeader.Height,
		"sender":        serverNodePeerId,
		"receiver":      clientPeerId,
	}

	m.logger.Info().Fields(fields).Msgf("Received GetBlockResponse: %s", blockRes)

	return nil
}

func (m *stateSync) GetSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse {
	return m.syncMetadataBuffer
}

// TODO(#352): Implement this function, currently a placeholder.
// Requests blocks one by one from its peers.
func (m *stateSync) StartSynching() error {
	current_height := m.GetBus().GetConsensusModule().CurrentHeight()
	var lastPersistedBlockHeight uint64

	if current_height == 0 {
		lastPersistedBlockHeight = 0
	} else {
		lastPersistedBlockHeight = current_height - 1
	}

	m.logger.Debug().Msgf("Starting synching, last persisted block %d, aggregated maxHeight %d", lastPersistedBlockHeight, m.aggregatedSyncMetadata.MaxHeight)

	// ADD BUSINESS LOGIC

	return nil
}

// TODO(#352): Implement this function, currently a placeholder.
// Returns max block height metadainfo received from all peers by aggregating responses in the buffer.
func (m *stateSync) aggregateMetadataResponses() *typesCons.StateSyncMetadataResponse {
	m.m.Lock()
	defer m.m.Unlock()

	metadataResponse := m.aggregatedSyncMetadata

	return metadataResponse
}

// TODO(#352): Implement this function, currently a placeholder.
// Periodically (initially by using timers) queries the network by sending metadata requests to peers using broadCastStateSyncMessage() function.
// Update frequency can be tuned accordingly to the state. Initially, it will have a default  behaviour.
func (m *stateSync) periodicMetaDataSynch() error {

	// set a timer to periodically query the network
	// form a metadata request
	// send to peers using broadCastStateSyncMessage()

	return nil
}
