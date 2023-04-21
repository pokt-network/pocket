package state_sync

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName             = "stateSyncModule"
	committedBlockHeightChannelSize = 100
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	Set(aggregatedMetaData *typesCons.StateSyncMetadataResponse)
	CommittedBlock(uint64)
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus                         modules.Bus
	logger                      *modules.Logger
	validators                  []*coreTypes.Actor
	aggregatedMetaData          *typesCons.StateSyncMetadataResponse
	committedBlockHeightChannel chan uint64
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (m *stateSync) CommittedBlock(height uint64) {
	m.committedBlockHeightChannel <- height
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	m.committedBlockHeightChannel = make(chan uint64, committedBlockHeightChannelSize)

	return m, nil
}

func (m *stateSync) Set(aggregatedMetaData *typesCons.StateSyncMetadataResponse) {
	m.logger.Info().Msg("State Sync Module Set")
	m.aggregatedMetaData = aggregatedMetaData
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
	m.validators, err = readCtx.GetAllValidators(int64(currentHeight))
	if err != nil {
		return err
	}

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
		// use raintree broadcast
		for _, val := range m.validators {
			if err := m.sendStateSyncMessage(stateSyncGetBlockMessage, cryptoPocket.AddressFromString(val.GetAddress())); err != nil {
				return err
			}
		}

		// wait for the block to be received and committed by consensus module
		receivedBlockHeight := <-m.committedBlockHeightChannel
		// TODO!: do we need to do this check? It should not happen
		if receivedBlockHeight != consensusMod.CurrentHeight() {
			return fmt.Errorf("received block height %d is not equal to current height %d", receivedBlockHeight, currentHeight)
		}
		//timer to check if block is received and committed
		currentHeight = consensusMod.CurrentHeight()
	}
	// syncing is complete, stop the state sync module
	return m.Stop()
}

// TODO(#352): check if node is a valdiator, if not send Consensus_IsSyncedNonValidator event
// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) Stop() error {
	m.logger.Info().Msg("Stop state sync moudule")
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator)
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
