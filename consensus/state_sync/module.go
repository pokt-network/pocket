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
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
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

	SyncStateSync() error
	HandleStateSyncBlockCommittedEvent(message *anypb.Any) error
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error
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

// Start performs state sync
// processes and aggregates all metadata collected in metadataReceived channel,
// requests missing blocks starting from its current height to the aggregated metadata's maxHeight,
// once the requested block is received and committed by consensus module, sends the next request for the next block,
// when all blocks are received and committed, stops the state sync process by calling its `Stop()` function.
func (m *stateSync) SyncStateSync() error {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()

	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return err
	}
	defer readCtx.Release()

	// TECHDEBT: We want to request blocks from all peers (staked or not) as opposed to just validators
	validators, err := readCtx.GetAllValidators(int64(currentHeight))
	if err != nil {
		return err
	}

	// Understand the view of the network
	aggregatedMetaData := m.getAggregatedStateSyncMetadata()
	maxHeight := aggregatedMetaData.MaxHeight

	// requests blocks from the current height to the aggregated metadata height
	for currentHeight <= maxHeight {
		m.logger.Info().Msgf("Sync is requesting block: %d, ending height: %d", currentHeight, maxHeight)

		// form the get block request message
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      currentHeight,
				},
			},
		}

		// Broadcast the get block request message from all the available peers on the network
		// TODO: Use P2P.broadcast instead of looping over the validators and sending the message to each one
		for _, val := range validators {
			if err := m.sendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress())); err != nil {
				return err
			}
		}

		// Wait for the consensus module to commit the requested block
		// If the block is not committed within some time, try re-requesting the block
		select {
		case blockHeight := <-m.committedBlocksChannel:
			// requested block is received and committed, continue to request the next block from the current height
			m.logger.Info().Msgf("Block %d is committed!", blockHeight)
		case <-time.After(blockWaitingPeriod):
			m.logger.Warn().Msgf("Timed out waiting for block %d to be committed...", currentHeight)
		}

		// Update the height and continue catching up to the latest known state
		currentHeight = consensusMod.CurrentHeight()
	}
	// syncing is complete and all requested blocks are committed, stop the state sync module
	return m.pauseSynching()
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

// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) pauseSynching() error {
	currentHeight := m.bus.GetConsensusModule().CurrentHeight()
	nodeAddress := m.bus.GetConsensusModule().GetNodeAddress()

	readCtx, err := m.bus.GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return err
	}
	defer readCtx.Release()

	nodeAddressBz, err := hex.DecodeString(nodeAddress)
	if err != nil {
		return err
	}
	isValidator, err := readCtx.GetValidatorExists(nodeAddressBz, int64(currentHeight))
	if err != nil {
		return err
	}

	if isValidator {
		return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
	}
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator)
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

// metadataSyncLoop periodically sends metadata requests to its peers to aggregate metadata related to synching the state.
// It is intended to be run as a background process via `go metadataSyncLoop`
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

	currentHeight := m.bus.GetConsensusModule().CurrentHeight()
	// TECHDEBT: This should be sent to all peers (full nodes, servicers, etc...), not just validators
	validators, err := m.getValidatorsAtHeight(currentHeight)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrPersistenceGetAllValidators.Error())
	}

	for _, val := range validators {
		anyMsg, err := anypb.New(stateSyncMetadataReqMessage)
		if err != nil {
			return err
		}
		// TECHDEBT: Revisit why we're not using `Broadcast` here instead of `Send`.
		if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(val.GetAddress()), anyMsg); err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrSendMessage.Error())
			return err
		}
	}

	return nil
}
