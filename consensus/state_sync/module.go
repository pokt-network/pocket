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
	//StartSyncing()
	HandleStateSyncBlockCommittedEvent(message *anypb.Any) error
	// active state sync
	CatchToHeight()
	SetActiveSyncHeight(height uint64)
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
	activeSyncHeight       uint64
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (m *stateSync) HandleStateSyncBlockCommittedEvent(event *anypb.Any) error {
	fmt.Println("newHeightEvent: ", event)
	evt, err := codec.GetCodec().FromAny(event)
	if err != nil {
		return err
	}

	switch event.MessageName() {

	case messaging.StateSyncBlockCommittedEventType:
		newCommitBlockEvent, ok := evt.(*typesCons.StateSyncBlockCommittedEvent)
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

func (m *stateSync) SetActiveSyncHeight(height uint64) {
	if height > m.activeSyncHeight {
		fmt.Println("SETTING from", m.activeSyncHeight, "to", height, "height")
		m.activeSyncHeight = height
	}
}

// func (m *stateSync) StartSyncing() {
// 	err := m.Start()
// 	if err != nil {
// 		m.logger.Error().Err(err).Msg("couldn't start state sync")
// 	}
// }

// Start performs passive state sync process, starting from the consensus module's current height to the aggragated metadata height
func (m *stateSync) Start() error {
	go m.metadataSyncLoop()
	go m.blockRequestLoop()

	return nil
}

// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) Stop() error {
	// check if the node is a validator
	// currentHeight := m.bus.GetConsensusModule().CurrentHeight()
	// nodeAddress := m.bus.GetConsensusModule().GetNodeAddress()
	// isValidator, err := m.bus.GetPersistenceModule().IsValidator(int64(currentHeight), nodeAddress)

	// if err != nil {
	// 	return err
	// }
	// m.logger.Info().Msg("Syncing is complete!")

	// if isValidator {
	// 	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
	// }
	// return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator)

	// TODO! check what else needs to be done here
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

// TODO(#352): Implement this function, currently a placeholder.
// metadataSyncLoop periodically sends metadata requests to its peers
// it is intended to be run as a background process
func (m *stateSync) metadataSyncLoop() {
	// runs as a background process
	// requests metadata from peers
	// sends received metadata to the metadataReceived channel
}

func (m *stateSync) blockRequestLoop() {
	// runs as a background process
	// requests blocks from the current height to the aggregated metadata height
	// sends received blocks to the blockReceived channel
}

// CatchToHeight performs active state sync
// TODO! check what to do with returning error
func (m *stateSync) CatchToHeight() {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		m.logger.Err(err).Msg("failed to create read context for persistence module")
		return
	}
	defer readCtx.Release()

	//get the current validators
	validators, err := readCtx.GetAllValidators(int64(currentHeight))
	if err != nil {
		m.logger.Err(err).Msg("failed to get all validators from persistence module")
		return
	}

	// requests blocks from the current height to the aggregated metadata height
	for currentHeight <= m.activeSyncHeight {
		m.logger.Info().Msgf("Sync is requesting block: %d, ending height: %d", currentHeight, m.activeSyncHeight)

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
		for _, val := range validators {
			if err := m.sendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress())); err != nil {
				m.logger.Err(err).Msg("failed to send state sync message")
				return
			}
		}

		// wait for the requested block to be received and committed by consensus module
		<-m.committedBlocksChannel

		// requested block is received and committed, continue to the next block from the current height
		currentHeight = consensusMod.CurrentHeight()
		fmt.Println("Current Height is Now: ", currentHeight)
	}

	// syncing is complete and all requested blocks are committed send validator synced event to the FSM
	// isValidator, err := m.bus.GetPersistenceModule().IsValidator(int64(currentHeight), nodeAddress)
	// if err != nil {
	// 	m.logger.Err(err).Msg("failed to check if the node is a validator")
	// }

	m.logger.Info().Msg("Active syncing is complete!")

	// var event coreTypes.StateMachineEvent
	// if isValidator {
	//  event := coreTypes.StateMachineEvent_Consensus_IsSyncedValidator
	// } else {
	// 	 event = coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator
	// }

	if err = m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator); err != nil {
		m.logger.Err(err).Msg("failed to send IsSyncedValidator event to the FSM")
	}
}
