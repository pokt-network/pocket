package state_sync

import (
	"fmt"
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

	IsServerModEnabled() bool
	EnableServerMode()
	SetLogPrefix(string)
}

var (
	_ modules.Module        = &stateSync{}
	_ StateSyncServerModule = &stateSync{}
)

type stateSync struct {
	bus modules.Bus

	currentMode SyncMode
	serverMode  bool

	logPrefix string
}

func CreateStateSync(bus modules.Bus) (modules.Module, error) {
	var m stateSync
	return m.Create(bus)
}

func (*stateSync) Create(bus modules.Bus) (modules.Module, error) {
	m := &stateSync{}
	bus.RegisterModule(m)

	// TODO: think about what must be the default mode,
	// Synched seems reasonable, as switching to pacemaker and sync modes must trigger operations
	// And target state for the node must be synched.
	m.currentMode = Synched

	m.serverMode = false

	return m, nil
}

func (m *stateSync) Start() error {
	return nil
}

func (m *stateSync) Stop() error {
	return nil
}

func (m *stateSync) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *stateSync) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *stateSync) GetModuleName() string {
	return stateSyncModuleName
}

func (m *stateSync) IsServerModEnabled() bool {
	return m.serverMode
}

func (m *stateSync) SetLogPrefix(logPrefix string) {
	m.logPrefix = logPrefix
}

func (m *stateSync) EnableServerMode() {
	m.currentMode = Server
	m.serverMode = true
}

func (m *stateSync) HandleGetBlockResponse(blockRes *typesCons.GetBlockResponse) error {
	m.nodeLog(fmt.Sprintf("Received get block response: %s", blockRes.Block.String()))
	return nil
}

func (m *stateSync) HandleStateSyncMetadataResponse(metaDataRes *typesCons.StateSyncMetadataResponse) error {
	m.nodeLog(fmt.Sprintf("Received get metadata response: %s", metaDataRes.String()))
	return nil
}

// IMPROVE: Remove this once we have a proper logging system.
func (m *stateSync) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.GetBus().GetConsensusModule().GetNodeId(), s)
}

// IMPROVE: Remove this once we have a proper logging system.
func (m *stateSync) nodeLogError(s string, err error) {
	log.Printf("[ERROR][%s][%d] %s: %v\n", m.logPrefix, m.GetBus().GetConsensusModule().GetNodeId(), s, err)
}
