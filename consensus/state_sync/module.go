package state_sync

import (
	"context"
	"sync"
	"time"

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

	// Handle a metadata response from a peer so this node can update its local view of the state
	// sync metadata available from its peers
	//HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error

	// Handle a block response from a peer so this node can update apply it to its local state
	// and catch up to the global world state
	//HandleGetBlockResponse(*typesCons.GetBlockResponse) error

	IsServerModEnabled() bool
	EnableServerMode() error
	DisableServerMode() error

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, nodeAddress cryptoPocket.Address, height uint64) error
	//IsSynched() bool
	//AggregateMetadataResponses() error
	SetSyncMetadataBuffer([]*typesCons.StateSyncMetadataResponse)
	GetSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse

	GetAggregatedSyncMetadata() *typesCons.StateSyncMetadataResponse

	StartSynching() error
}

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

	// node will be periodically checking if its up to date.
	// and it will be updating the AggregatedSynchMetaData.
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
	m.logger.Debug().Msgf("GETTING AGGREGATED SYNC METADATA TO: %s, nodeId: %d", m.aggregatedSyncMetadata, m.GetBus().GetConsensusModule().GetNodeId())
	return m.aggregatedSyncMetadata
}

func (m *stateSync) SetAggregatedSyncMetadata(metaData *typesCons.StateSyncMetadataResponse) {
	m.logger.Debug().Msgf("SETTING AGGREGATED SYNC METADATA TO: %s, nodeId: %d", metaData, m.GetBus().GetConsensusModule().GetNodeId())
	m.aggregatedSyncMetadata = metaData
}

func (m *stateSync) SetSyncMetadataBuffer(aggregatedSyncMetadata []*typesCons.StateSyncMetadataResponse) {
	m.m.Lock()
	defer m.m.Unlock()
	m.syncMetadataBuffer = aggregatedSyncMetadata
}

func (m *stateSync) GetSyncMetadataBuffer() []*typesCons.StateSyncMetadataResponse {
	return m.syncMetadataBuffer
}

// TODO! implement this function, placeholder
// This function requests blocks one by one from peers
func (m *stateSync) StartSynching() error {
	m.logger.Debug().Msgf("StartSynching() called NEWGOKHAN")

	current_height := m.GetBus().GetConsensusModule().CurrentHeight()
	//! TODO CHECK THIS
	lastPersistedBlockHeight := current_height - 1
	m.logger.Debug().Msgf("Last persisted block %d, Aggregated maxHeight %d", lastPersistedBlockHeight, m.aggregatedSyncMetadata.MaxHeight)

	for i := lastPersistedBlockHeight; i < m.aggregatedSyncMetadata.MaxHeight; i++ {
		m.logger.Debug().Msgf("StartSynching() Requesting block %d", i)
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
					Height:      i,
				},
			},
		}

		m.broadCastStateSyncMessage(stateSyncGetBlockMessage, current_height)
	}

	return nil
}

// Returns max block height metadainfo received from all peers.
// This function requests blocks one by one from peers thorughusing p2p module request, aggregates responses.
// It requests blocks one by one from peers thorughusing p2p module request
func (m *stateSync) aggregateMetadataResponses() *typesCons.StateSyncMetadataResponse {
	m.m.Lock()
	defer m.m.Unlock()

	metadataResponse := m.aggregatedSyncMetadata

	//aggregate metadataResponses by setting the metadataResponse
	for _, meta := range m.syncMetadataBuffer {
		if meta.MaxHeight > metadataResponse.MaxHeight {
			m.logger.Debug().Msgf("YOYOOYYO (): %s", metadataResponse)
			metadataResponse.MaxHeight = meta.MaxHeight
		}

		if meta.MinHeight < metadataResponse.MinHeight {
			metadataResponse.MinHeight = meta.MinHeight
		}
	}

	m.logger.Debug().Msgf("GOKHAN aggregateMetadataResponses, max height: %d", metadataResponse.MaxHeight)

	//clear the buffer
	m.syncMetadataBuffer = make([]*typesCons.StateSyncMetadataResponse, 0)

	return metadataResponse
}

// This function periodically checks if node is up to date with the network by sending metadata requests to peers.
// It updates the aggregatedSyncMetadata field.
// This update frequency can be tuned accordingly to the state. Currently, it has a default  behaviour.
func (m *stateSync) periodicMetaDataSynch() error {
	m.logger.Debug().Msgf("periodicSynch() called GOKHAN")

	//add timer channel with context to cancel the timer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// // form a metaData request
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.logger.Debug().Msg("Periodic metadata synch is triggered")
			currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()

			//broadcast metadata request to all peers
			//err := m.synch(currentHeight)
			err := m.broadCastStateSyncMessage(stateSyncMetaDataReqMessage, currentHeight)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return nil

		}
	}
}
