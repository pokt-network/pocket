package state_sync

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName = "stateSyncModule"
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	SendStateSyncMessage(msg *typesCons.StateSyncMessage, nodeAddress cryptoPocket.Address, height uint64) error
	StateSyncLogHelper(receiverPeerAddress string) map[string]any
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncModule       = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus    modules.Bus
	logger *modules.Logger

	logPrefix string
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

// TODO (#352): implement this function
// Start performs state sync
func (m *stateSync) Start() error {
	// gets and aggregated received metadata in metadataReceived channel of consensus module,
	// requests blocks starting from currentHeight to maxHeight via `broadcastStateSyncMessage()` function,
	// waits received and committed block event in `blockReceived` channel for each block,
	// once the event is received `blockReceived` channel, requests the next block.
	// when all blocks are received and committed, stops the state sync process, by colling `m.Stop()` function.

	return nil
}

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

func (m *stateSync) SetLogPrefix(logPrefix string) {
	m.logPrefix = logPrefix
}
