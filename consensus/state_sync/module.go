package state_sync

import (
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName = "stateSyncModule"
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus    modules.Bus
	logger *modules.Logger
}

func CreateStateSync(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(stateSync).Create(bus, options...)
}

func (*stateSync) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &stateSync{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	return m, nil
}

// TODO(#352): implement this function
// Start performs state sync
func (m *stateSync) Start() error {
	// processes and aggregates all metadata collected in metadataReceived channel,
	// requests missing blocks starting from its current height to the aggregated metadata's maxHeight,
	// once the requested block is received and committed by consensus module, sends the next request for the next block,
	// when all blocks are received and committed, stops the state sync process by calling its `Stop()` function.
	return nil
}

// TODO(#352): check if node is a valdiator, if not send Consensus_IsSyncedNonValidator event
// Stop stops the state sync process, and sends `Consensus_IsSyncedValidator` FSM event
func (m *stateSync) Stop() error {

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
