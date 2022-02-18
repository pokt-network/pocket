package statesync

import (
	"fmt"
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"
)

type SyncState uint8

const (
	Unknown      SyncState = iota
	Synched                // Blockchain is synched and participating in consensus.
	BlockSync              // TODO: SlowSync: Synching blockchain block by block.
	SnapshotSync           // TODO: FastSync: synching blockchain from a snapshot.
)

// TODO: Sync with Otto/Hamza on how this should be implemented.
type StateSyncModule interface {
	modules.Module

	IsSynched() bool
}

type stateSyncModule struct {
	StateSyncModule
	pocketBusMod modules.Bus
	syncState    SyncState
}

func Create(
	cfg *config.Config,
) (StateSyncModule, error) {
	return &stateSyncModule{
		syncState: Unknown,
	}, nil
}

func (m *stateSyncModule) Start() error {
	m.syncState = Unknown

	// Need to get block hash from PersistenceContext
	//prevHeight := uint64(height) - 1
	blockHeight := uint64(1)
	err := error(nil)
	//prevBlockHash, err := m.GetBus().GetPersistenceModule().Get GetpersistenceModule().GetBlockHash()

	//blockHeight, err := m.GetBus().GetPersistenceModule(). GetpersistenceModule().GetLatestBlockHeight()
	if err == nil {
		log.Println("[WARN] Persisted block data not found, synching from genesis.")
		m.syncFromGenesis()
	} else {
		log.Println("[WARN] Starting sync from block height: ", blockHeight)
		m.syncFromHeight(blockHeight)
	}

	return nil
}

func (m *stateSyncModule) Stop() error {
	return fmt.Errorf("StateSyncModule.Stop Not implemented")
}

func (m *stateSyncModule) IsSynched() bool {
	return m.syncState == Synched
}

func (m *stateSyncModule) syncFromGenesis() {
	// TODO
}

func (m *stateSyncModule) syncFromHeight(blockHeight uint64) {
	log.Println("[TODO] Implement sync from height...")
	m.syncFromGenesis()
}
