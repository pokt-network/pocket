package state_sync

import (
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
	stateSyncModuleName       = "stateSyncModule"
	committedBlocsChannelSize = 100
	blockWaitingPeriod        = 30 * time.Second
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	SetAggregatedMetadata(aggregatedMetaData *typesCons.StateSyncMetadataResponse)
	StartSyncing()
	HandleStateSyncBlockCommittedEvent(message *anypb.Any) error
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus                    modules.Bus
	logger                 *modules.Logger
	aggregatedMetaData     *typesCons.StateSyncMetadataResponse
	committedBlocksChannel chan uint64
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
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

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())
	m.committedBlocksChannel = make(chan uint64, committedBlocsChannelSize)

	return m, nil
}

func (m *stateSync) SetAggregatedMetadata(aggregatedMetaData *typesCons.StateSyncMetadataResponse) {
	m.aggregatedMetaData = aggregatedMetaData
}

// TODO: Remove this. This function added as a hack to be able to check return value of m.Start().
func (m *stateSync) StartSyncing() {
	err := m.Start()
	if err != nil {
		m.logger.Error().Err(err).Msg("couldn't start state sync")
	}
}

// Start performs state sync
// processes and aggregates all metadata collected in metadataReceived channel,
// requests missing blocks starting from its current height to the aggregated metadata's maxHeight,
// once the requested block is received and committed by consensus module, sends the next request for the next block,
// when all blocks are received and committed, stops the state sync process by calling its `Stop()` function.
func (m *stateSync) Start() error {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()

	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return err
	}
	defer readCtx.Release()

	//get the current validators
	validators, err := readCtx.GetAllValidators(int64(currentHeight))
	if err != nil {
		return err
	}

	// requests blocks from the current height to the aggregated metadata height
	for currentHeight <= m.aggregatedMetaData.MaxHeight {
		m.logger.Info().Msgf("Sync is requesting block: %d, ending height: %d", currentHeight, m.aggregatedMetaData.MaxHeight)

		// form the get block request message
		stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
			Message: &typesCons.StateSyncMessage_GetBlockReq{
				GetBlockReq: &typesCons.GetBlockRequest{
					PeerAddress: nodeAddress,
					Height:      currentHeight,
				},
			},
		}

		// broadcast the get block request message to all validators
		// TODO: use raintree broadcast
		for _, val := range validators {
			if err := m.sendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress())); err != nil {
				return err
			}
		}

		// wait to receive requested block for blockWaitingPeriod. If the block is received next block will be requested, otherwise the current block will be requested again
		select {
		case blockHeight := <-m.committedBlocksChannel:
			// requested block is received and committed, continue to request the next block from the current height
			m.logger.Info().Msgf("Block %d is committed!", blockHeight)
		case <-time.After(blockWaitingPeriod):
		}

		// wait for the requested block to be received and committed by consensus module
		//<-m.committedBlocksChannel

		currentHeight = consensusMod.CurrentHeight()
	}
	// syncing is complete and all requested blocks are committed, stop the state sync module
	return m.Stop()
}

// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) Stop() error {
	// check if the node is a validator
	currentHeight := m.bus.GetConsensusModule().CurrentHeight()
	nodeAddress := m.bus.GetConsensusModule().GetNodeAddress()

	readCtx, err := m.bus.GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return err
	}
	defer readCtx.Release()

	isValidator, err := readCtx.IsValidator(int64(currentHeight), nodeAddress)
	if err != nil {
		return err
	}

	if isValidator {
		return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
	}
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator)
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
