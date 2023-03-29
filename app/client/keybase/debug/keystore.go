//go:build debug

package debug

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/build"
	"github.com/pokt-network/pocket/logger"
)

const debugKeybaseSuffix = "/.pocket/keys"

var (
	// TODO: Allow users to override this value via `datadir` flag or env var or config file
	debugKeybasePath string
)

// Initialise the debug keybase with the 999 validator keys from the private-keys manifest file
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Cannot find user home directory")
	}
	debugKeybasePath = homeDir + debugKeybaseSuffix

	// Initialise the debug keybase with the 999 validators
	if err := initializeDebugKeybase(); err != nil {
		logger.Global.Fatal().Err(err).Msg("Cannot initialise the keybase with the validator keys")
	}
}

func initializeDebugKeybase() error {
	// Create/Open the keybase at `$HOME/.pocket/keys`
	kb, err := keybase.NewBadgerKeybase(debugKeybasePath)
	if err != nil {
		return err
	}
	db, err := kb.GetBadgerDB()
	if err != nil {
		return err
	}

	if err := restoreBadgerDB(build.DebugKeybaseBackup, db); err != nil {
		return err
	}

	// Close DB connection
	if err := kb.Stop(); err != nil {
		return err
	}

	return nil
}

func restoreBadgerDB(backupData []byte, db *badger.DB) error {
	logger.Global.Debug().Msg("Debug keybase initializing... Restoring from the embedded backup file...")

	// Create a temporary directory to store the backup data
	tempDir, err := ioutil.TempDir("", "badgerdb-restore")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Write the backup data to a file in the temporary directory
	backupFilePath := filepath.Join(tempDir, "backup")
	if err := ioutil.WriteFile(backupFilePath, backupData, 0644); err != nil {
		return err
	}

	backupFile, err := os.Open(backupFilePath)
	if err != nil {
		return err
	}
	defer backupFile.Close()

	if err := db.Load(backupFile, 4); err != nil {
		return err
	}

	return nil
}
