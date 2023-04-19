package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName       = "stateSyncModule"
	committedBlocsChannelSize = 100
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	Set(aggregatedMetaData *typesCons.StateSyncMetadataResponse)
	CommittedBlock(uint64)
	StartSyncing()
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus                    modules.Bus
	logger                 *modules.Logger
	validators             []*coreTypes.Actor
	aggregatedMetaData     *typesCons.StateSyncMetadataResponse
	committedBlocksChannel chan uint64
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

// CommittedBlock is called by the consensus module when a block received by the network is committed by blockApplicationLoop() function
func (m *stateSync) CommittedBlock(height uint64) {
	m.committedBlocksChannel <- height
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

func (m *stateSync) Set(aggregatedMetaData *typesCons.StateSyncMetadataResponse) {
	m.aggregatedMetaData = aggregatedMetaData
}

func (m *stateSync) StartSyncing() {
	err := m.Start()
	if err != nil {
		m.logger.Error().Err(err).Msg("couldn't start state sync")
	}
}

// Start performs state sync process, starting from the consensus module's current height to the aggragated metadata height
func (m *stateSync) Start() error {
	consensusMod := m.bus.GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	nodeAddress := consensusMod.GetNodeAddress()
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(int64(currentHeight))
	if err != nil {
		return err
	}
	defer readCtx.Release()

	// get the current validators
	m.validators, err = readCtx.GetAllValidators(int64(currentHeight))
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
		for _, val := range m.validators {
			if err := m.sendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress())); err != nil {
				return err
			}
		}

		// wait for the requested block to be received and committed by consensus module
		<-m.committedBlocksChannel

		// requested block is received and committed, continue to the next block from the current height
		currentHeight = consensusMod.CurrentHeight()
	}
	// syncing is complete and all requested blocks are committed, stop the state sync module
	return m.Stop()
}

// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) Stop() error {
	// check if the node is a validator
	isValidator, err := m.bus.GetConsensusModule().IsValidator()
	if err != nil {
		return err
	}
	m.logger.Info().Msg("Syncing is complete!")

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
