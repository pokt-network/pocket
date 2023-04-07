package savepoints

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	moduleName = "savepoint_factory"
)

var _ modules.SavepointFactory = &savepointFactory{}

type savepointFactory struct {
	readContext modules.PersistenceReadContext
	logger      *modules.Logger

	m sync.Mutex
}

func NewSavepointFactory(readContext modules.PersistenceReadContext) modules.SavepointFactory {
	return &savepointFactory{
		readContext: readContext,
		logger:      logger.Global.CreateLoggerForModule(moduleName),
	}
}

func (sm *savepointFactory) CreateSavepoint(height int64) (modules.PersistenceReadContext, error) {
	log := sm.logger.With().Fields(map[string]interface{}{
		"height": height,
		"source": "CreateSavepoint",
	}).Logger()
	log.Debug().Msg("Creating savepoint...")

	sm.m.Lock()
	defer sm.m.Unlock()

	currentHeight, err := sm.readContext.GetHeight()
	if err != nil {
		return nil, err
	}
	if currentHeight != height {
		return nil, fmt.Errorf("savepoint height %d does not match read context height %d", height, currentHeight)
	}

	nodeStores, valueStores := sm.readContext.GetKVStores()

	nodeStoresTempDir, err := ioutil.TempDir("", "pocketv1-snapshot-nodestores-*")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp directory for node stores")
	}
	valueStoresTempDir, err := ioutil.TempDir("", "pocketv1-snapshot-valuestores-*")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp directory for value stores")
	}

	log.Debug().
		Str("nodeStoresTempDir", nodeStoresTempDir).
		Str("valueStoresTempDir", valueStoresTempDir).
		Msg("Temp folders created")

	for treeId, nodeStore := range nodeStores {
		wg := new(sync.WaitGroup)
		wg.Add(2)

		go backupKVStore(nodeStore, nodeStoresTempDir, wg)
		go backupKVStore(valueStores[treeId], valueStoresTempDir, wg)

		wg.Wait()
	}

	// TODO: here we have to rehydrate the backups of the stores into the savepoint in a way that allows us to query them...
	// without SQL help

	return &savepoint{
		height:          height,
		nodeStoresPath:  nodeStoresTempDir,
		valueStoresPath: valueStoresTempDir,
	}, nil
}

func backupKVStore(kvStore kvstore.BackupableKVStore, backupDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	storeName := kvStore.GetName()
	targetFilePath := filepath.Join(backupDir, fmt.Sprintf("%s.bak", storeName))

	backupFile, err := os.Create(targetFilePath)
	if err != nil {
		log.Fatal().Err(err).Str("storeName", storeName).Msg("Failed to create backup file")
	}
	defer backupFile.Close()

	log.Debug().Str("storeName", storeName).Str("targetFilePath", targetFilePath).Msg("Backing up store")
	if _, err := kvStore.Backup(backupFile, 0); err != nil {
		log.Fatal().Err(err).Str("storeName", storeName).Msg("Failed to backup store")
	}
}
