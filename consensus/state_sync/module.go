package state_sync

import (
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

	// Handle a metadata response from a peer so this node can update its local view of the state
	// sync metadata available from its peers
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error

	// Handle a block response from a peer so this node can update apply it to its local state
	// and catch up to the global world state
	HandleGetBlockResponse(*typesCons.GetBlockResponse) error

	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, nodeAddress cryptoPocket.Address, height uint64) error
	IsOutOfSync() bool
	AggregateMetadataResponses() error
	GetAggregatedSyncMetadata() *typesCons.StateSyncMetadataResponse
	Sync() error
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus    modules.Bus
	logger *modules.Logger

	logPrefix  string
	serverMode bool

	aggregatedSyncMetadata *typesCons.StateSyncMetadataResponse
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

	return m, nil
}

func (m *stateSync) Start() error {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

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

// TODO(#352): Implement this function
// Placeholder function
func (m *stateSync) HandleGetBlockResponse(blockRes *typesCons.GetBlockResponse) error {
	consensusMod := m.GetBus().GetConsensusModule()
	serverNodePeerId := consensusMod.GetNodeAddress()
	clientPeerId := blockRes.PeerAddress

	fields := map[string]any{
		"currentHeight": blockRes.Block.BlockHeader.Height,
		"sender":        serverNodePeerId,
		"receiver":      clientPeerId,
	}

	m.logger.Info().Fields(fields).Msgf("Received GetBlockResponse: %s", blockRes)

	return nil
}

// TODO(#352): Implement the business to handle these correctly
// Placeholder function
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

	return nil
}

func (m *stateSync) GetAggregatedSyncMetadata() *typesCons.StateSyncMetadataResponse {
	return m.aggregatedSyncMetadata
}

// TODO! implement this function, placeholder
// Returns max block height metadainfo received from all peers.
// This function requests blocks one by one from peers thorughusing p2p module request, aggregates responses
func (m *stateSync) AggregateMetadataResponses() error {
	metadataResponse := &typesCons.StateSyncMetadataResponse{}
	m.aggregatedSyncMetadata = metadataResponse

	return nil
}

// TODO! implement this function, placeholder
func (m *stateSync) IsOutOfSync() bool {
	err := m.AggregateMetadataResponses()
	if err != nil {
		// TODO: correct error handling
		m.logger.Error().Err(err)
		return false
	}

	return m.GetBus().GetConsensusModule().CurrentHeight()-1 < m.aggregatedSyncMetadata.MaxHeight-1
}

// TODO! implement this function, placeholder
// This function requests blocks one by one from peers thorughusing p2p module request
func (m *stateSync) Sync() error {

	return nil
}
