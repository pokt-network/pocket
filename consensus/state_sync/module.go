package state_sync

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	stateSyncModuleName = "stateSyncModule"
)

type Sync_Mode string

const (
	Snyc      Sync_Mode = "sync"
	Synched   Sync_Mode = "synched"
	Pacemaker Sync_Mode = "pacemaker"
	Server    Sync_Mode = "server"
)

type StateSyncModule interface {
	modules.Module
	StateSyncServerModule

	// Handle a metadata response from a peer so this node can update its local view of the state
	// sync metadata available from its peers
	HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error

	// Handle a block response from a peer so this node can update apply it to its local state
	// and catch up to the global world state
	HandleGetBlockResponse(*typesCons.GetBlockResponse) error

	//HandleStateSyncMetadataRequest

	IsServerModEnabled() bool
	EnableServerMode()
}

var (
	_ modules.Module        = &stateSyncModule{}
	_ StateSyncServerModule = &stateSyncModule{}
)

type stateSyncModule struct {
	bus modules.Bus

	currentMode Sync_Mode
	serverMode  bool
}

func CreateStateSync(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m stateSyncModule
	return m.Create(runtimeMgr)
}

func (*stateSyncModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	//! TODO: think about what must be the default state?
	return &stateSyncModule{
		bus:         nil,
		currentMode: Synched,
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
	m.currentMode = Server
	m.serverMode = true
}

func (m *stateSyncModule) HandleGetBlockResponse(*typesCons.GetBlockResponse) error {
	//! TODO implement
	return nil
}

func (m *stateSyncModule) HandleStateSyncMetadataResponse(*typesCons.StateSyncMetadataResponse) error {
	//! TODO implement
	return nil
}
