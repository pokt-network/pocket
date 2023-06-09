package state_sync

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	stateSyncModuleName = "stateSyncModule"
	// TODO: Make these configurable
	blockWaitingPeriod        = 30 * time.Second
	committedBlocsChannelSize = 100
	metadataChannelSize       = 1000
	blocksChannelSize         = 1000
	metadataSyncPeriod        = 45 * time.Second
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	HandleStateSyncBlockCommittedEvent(message *anypb.Any) error
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error

	// TECHDEBT: This function can be removed once the dependency of state sync on the FSM module is removed.
	StartSynchronousStateSync() error
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus    modules.Bus
	logger *modules.Logger

	// metadata responses received from peers are collected in this channel
	metadataReceived chan *typesCons.StateSyncMetadataResponse

	committedBlocksChannel chan uint64
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{
		metadataReceived:       make(chan *typesCons.StateSyncMetadataResponse, metadataChannelSize),
		committedBlocksChannel: make(chan uint64, committedBlocsChannelSize),
	}
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	for _, option := range options {
		option(m)
	}
	bus.RegisterModule(m)

	return m, nil
}

func (m *stateSync) Start() error {
	go m.metadataSyncLoop()
	return nil
}

// Start a synchronous state sync process to catch up to the network
// 1. Processes and aggregates all metadata collected in metadataReceived channel
// 2. Requests missing blocks until the maximum seen block is retrieved
// 3. Perform (2) one-by-one, applying and validating each block while doing so
// 4. Once all blocks are received and committed, stop the synchronous state sync process
func (m *stateSync) StartSynchronousStateSync() error {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()
	nodeAddressBz, err := hex.DecodeString(nodeAddress)
	if err != nil {
		return err
	}

	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return err
	}
	defer readCtx.Release()

	// Get a view into the state of the network
	_, maxHeight := m.getAggregatedStateSyncMetadata()

	// Synchronously request block requests from the current height to the aggregated metadata height
	for currentHeight <= maxHeight {
		m.logger.Info().Msgf("Sync is requesting block: %d, ending height: %d", currentHeight, maxHeight)

		// form the get block request message
		stateSyncGetBlockMsg := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      currentHeight,
				},
			},
		}
		anyProtoStateSyncMsg, err := anypb.New(stateSyncGetBlockMsg)
		if err != nil {
			return err
		}

		// Broadcast the block request
		if err := m.GetBus().GetP2PModule().Broadcast(anyProtoStateSyncMsg); err != nil {
			return err
		}

		// Wait for the consensus module to commit the requested block and re-try on timeout
		select {
		case blockHeight := <-m.committedBlocksChannel:
			m.logger.Info().Msgf("Block %d is committed!", blockHeight)
		case <-time.After(blockWaitingPeriod):
			m.logger.Warn().Msgf("Timed out waiting for block %d to be committed...", currentHeight)
		}

		// Update the height and continue catching up to the latest known state
		currentHeight = consensusMod.CurrentHeight()
	}

	// Checked if the synched node is a validator or not
	isValidator, err := readCtx.GetValidatorExists(nodeAddressBz, int64(currentHeight))
	if err != nil {
		return err
	}

	// Send out the appropriate FSM event now that the node is caught up
	if isValidator {
		return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
	}
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator)
}

func (m *stateSync) HandleStateSyncMetadataResponse(res *typesCons.StateSyncMetadataResponse) error {
	m.metadataReceived <- res
	return nil
}

func (m *stateSync) HandleStateSyncBlockCommittedEvent(event *anypb.Any) error {
	evt, err := codec.GetCodec().FromAny(event)
	if err != nil {
		return err
	}

	if event.MessageName() == messaging.StateSyncBlockCommittedEventType {
		newCommitBlockEvent, ok := evt.(*messaging.StateSyncBlockCommittedEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to StateSyncBlockCommittedEvent")
		}
		m.committedBlocksChannel <- newCommitBlockEvent.Height
	}
	return nil
}

func (m *stateSync) Stop() error {
	m.logger.Log().Msg("Draining and closing metadataReceived and blockResponse channels")
	for {
		select {
		case metaData, ok := <-m.metadataReceived:
			if ok {
				m.logger.Info().Msgf("Drained metadata message: %s", metaData)
			} else {
				close(m.metadataReceived)
				return nil
			}
		default:
			close(m.metadataReceived)
			return nil
		}
	}
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

// metadataSyncLoop periodically sends metadata requests to its peers to collect &
// aggregate metadata related to synching the state.
// It is intended to be run as a background process via a goroutine.
func (m *stateSync) metadataSyncLoop() error {
	logger := m.logger.With().Str("source", "metadataSyncLoop").Logger()
	ctx := context.TODO()

	ticker := time.NewTicker(metadataSyncPeriod)
	for {
		select {
		case <-ticker.C:
			logger.Info().Msg("Background metadata sync check triggered")
			if err := m.broadcastMetadataRequests(); err != nil {
				logger.Error().Err(err).Msg("Failed to send metadata requests")
				return err
			}

		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

// broadcastMetadataRequests sends a metadata request to all peers in the network to understand
// the state of the network and determine if the node is behind.
func (m *stateSync) broadcastMetadataRequests() error {
	stateSyncMetadataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: m.GetBus().GetConsensusModule().GetNodeAddress(),
			},
		},
	}
	anyMsg, err := anypb.New(stateSyncMetadataReqMessage)
	if err != nil {
		return err
	}
	if err := m.GetBus().GetP2PModule().Broadcast(anyMsg); err != nil {
		return err
	}
	return nil
}
