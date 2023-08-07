package state_sync

import (
	"context"
	"encoding/hex"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
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
	metadataSyncPeriod        = 10 * time.Second
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	HandleBlockCommittedEvent(*messaging.ConsensusNewHeightEvent)
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse)

	// TECHDEBT: This function can be removed once the dependency of state sync on the FSM module is removed.
	StartSynchronousStateSync()
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
func (m *stateSync) StartSynchronousStateSync() {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()
	nodeAddressBz, err := hex.DecodeString(nodeAddress)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to decode node address")
		return
	}

	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to create read context")
		return
	}
	defer readCtx.Release()

	// Get a view into the state of the network
	_, maxHeight := m.getAggregatedStateSyncMetadata()

	m.logger.Info().
		Uint64("current_height", currentHeight).
		Uint64("max_height", maxHeight).
		Msg("Synchronous state sync is requesting blocks...")

	// Synchronously request block requests from the current height to the aggregated metadata height
	// Note that we are using `<=` because:
	// - maxHeight is the max * committed * height of the network
	// - currentHeight is the latest * committing * height of the node

	// We do not need to request the genesis block from anyone
	if currentHeight == 0 {
		currentHeight += 1
	}

	for currentHeight <= maxHeight {
		m.logger.Info().Msgf("Synchronous state sync is requesting block: %d, ending height: %d", currentHeight, maxHeight)

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
			m.logger.Error().Err(err).Msg("Failed to create Any proto")
			return
		}

		// Broadcast the block request
		if err := m.GetBus().GetP2PModule().Broadcast(anyProtoStateSyncMsg); err != nil {
			m.logger.Error().Err(err).Msg("Failed to broadcast state sync message")
			return
		}

		// Wait for the consensus module to commit the requested block and re-try on timeout
		select {
		case blockHeight := <-m.committedBlocksChannel:
			m.logger.Info().Msgf("State sync received event that block %d is committed!", blockHeight)
		case <-time.After(blockWaitingPeriod):
			m.logger.Error().Msgf("Timed out waiting for block %d to be committed...", currentHeight)
		}

		// Update the height and continue catching up to the latest known state
		currentHeight = consensusMod.CurrentHeight()
	}

	// Checked if the synched node is a validator or not
	isValidator, err := readCtx.GetValidatorExists(nodeAddressBz, int64(currentHeight))
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to check if validator exists")
		return
	}

	// Send out the appropriate FSM event now that the node is caught up
	if isValidator {
		err = m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
	} else {
		err = m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator)
	}
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to send state machine event")
		return
	}
}

func (m *stateSync) HandleStateSyncMetadataResponse(res *typesCons.StateSyncMetadataResponse) {
	m.logger.Info().Fields(map[string]any{
		"peer_address": res.PeerAddress,
		"min_height":   res.MinHeight,
		"max_height":   res.MaxHeight,
	}).Msg("Handling state sync metadata response")
	m.metadataReceived <- res

	if res.MaxHeight > 0 && m.GetBus().GetConsensusModule().CurrentHeight() <= res.MaxHeight {
		if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsUnsynced); err != nil {
			m.logger.Error().Err(err).Msg("Failed to send state machine event")
		}
	}
}

func (m *stateSync) HandleBlockCommittedEvent(msg *messaging.ConsensusNewHeightEvent) {
	m.logger.Info().Msg("Handling state sync block committed event")
	m.committedBlocksChannel <- msg.Height
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
func (m *stateSync) metadataSyncLoop() {
	metaSyncLoopLogger := m.logger.With().Str("source", "metadataSyncLoop").Logger()
	ctx := context.TODO()

	ticker := time.NewTicker(metadataSyncPeriod)
	for {
		select {
		case <-ticker.C:
			metaSyncLoopLogger.Info().Msg("Background metadata sync check triggered")
			if err := m.broadcastMetadataRequests(); err != nil {
				metaSyncLoopLogger.Error().Err(err).Msg("Failed to send metadata requests")
			}

		case <-ctx.Done():
			ticker.Stop()
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
