package state_sync

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	DefaultLogPrefix    = "NODE"
	stateSyncModuleName = "stateSyncModule"
)

type SyncMode string

const (
	Snyc      SyncMode = "sync"
	Synched   SyncMode = "synched"
	Pacemaker SyncMode = "pacemaker"
	Server    SyncMode = "server"
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

	currentMode SyncMode
	serverMode  bool

	//REFACTOR: this should be removed, when we build a shared and proper logger
	logPrefix string
}

func CreateStateSync(bus modules.Bus) (modules.Module, error) {
	var m stateSyncModule
	return m.Create(bus)
}

func (*stateSyncModule) Create(bus modules.Bus) (modules.Module, error) {
	//! TODO: think about what must be the default mode
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

// IMPROVE: Remove this once we have a proper logging system.
func (m *stateSyncModule) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.GetBus().GetConsensusModule().GetNodeId(), s)
}

// IMPROVE: Remove this once we have a proper logging system.
func (m *stateSyncModule) nodeLogError(s string, err error) {
	log.Printf("[ERROR][%s][%d] %s: %v\n", m.logPrefix, m.GetBus().GetConsensusModule().GetNodeId(), s, err)
}
