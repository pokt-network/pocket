package state_sync

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName = "stateSyncModule"
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	// Handle a metadata response from a peer so this node can update its local view of the state
	// sync metadata available from its peers
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error

	// Handle a block response from a peer so this node can update apply it to its local state
	// and catch up to the global world state
	HandleStateSyncBlockResponse(*typesCons.StateSyncMetadataResponse) error

	IsServerModEnabled() bool
	EnableServerMode()
}

var (
	_ modules.Module        = &stateSyncModule{}
	_ StateSyncServerModule = &stateSyncModule{}
)

type stateSyncModule struct {
	bus modules.Bus

	currentMode string
	serverMode  bool
}

func CreateStateSync(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m StateSyncModule
	return m.Create(runtimeMgr)
}

func (*stateSyncModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return &stateSyncModule{
		bus:         nil,
		currentMode: "asd",
		serverMode:  false,
	}, nil
}

func (m *stateSyncModule) Start() error {
	return nil
}

func (m *stateSyncModule) Stop() error {
	return nil
}

func (m *stateSyncModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *stateSyncModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *stateSyncModule) GetModuleName() string {
	return stateSyncModuleName
}

func (m *stateSyncModule) IsServerModEnabled() bool {
	return m.serverMode
}

func (m *stateSyncModule) EnableServerMode() {
	m.serverMode = true
}
