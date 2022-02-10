package statesync

import (
	"fmt"
	"log"

	"pocket/consensus/pkg/shared/context"
	"pocket/consensus/pkg/shared/modules"
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
	modules.PocketModule

	IsSynched() bool
}

type stateSyncModule struct {
	*modules.BasePocketModule

	syncState SyncState
}

func Create(
	ctx *context.PocketContext,
	base *modules.BasePocketModule,
) (StateSyncModule, error) {
	return &stateSyncModule{
		BasePocketModule: base,

		syncState: Unknown,
	}, nil
}

func (m *stateSyncModule) Start(ctx *context.PocketContext) error {
	m.syncState = Unknown

	blockHeight, err := m.GetPocketBusMod().GetPersistanceModule().GetLatestBlockHeight()
	if err == nil {
		log.Println("[WARN] Persisted block data not found, synching from genesis.")
		m.syncFromGenesis()
	} else {
		log.Println("[WARN] Starting sync from block height: ", blockHeight)
		m.syncFromHeight(blockHeight)
	}

	return nil
}

func (m *stateSyncModule) Stop(ctx *context.PocketContext) error {
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
