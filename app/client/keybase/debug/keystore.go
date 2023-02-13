package debug

import (
	"errors"
	"fmt"
	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/shared/crypto"
	"gopkg.in/yaml.v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	validatorNumber    = 999
	debugKeybaseSuffix = "/.pocket/keys"
)

var (
	DebugKeybasePath string
)

// Initialise the debug keybase with the 999 validator keys from the private-keys manifest file
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("[ERROR] Cannot find user home directory: %s", err.Error())
	}
	DebugKeybasePath = homeDir + debugKeybaseSuffix

	if err := InitialiseDebugKeybase(); err != nil { // Initialise the debug keybase with the 999 validators
		log.Fatalf("[ERROR] Cannot initialise the keybase with the validator keys: %s", err.Error())
	}
}

// Creates/Opens the DB at `$HOME/.pocket/keys`
func NewDebugKeybase() (keybase.Keybase, error) {
	return keybase.NewKeybase(DebugKeybasePath)
}

// Struct to store the private-keys yaml file
type yamlConfig struct {
	ApiVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   map[string]string `yaml:"metadata"`
	Type       string            `yaml:"type"`
	StringData map[string]string `yaml:"stringData"`
}

// Creates/Opens the DB and initialises the keys from the YAML file
// FOR DEV/LOCANET PURPOSES ONLY
func InitialiseDebugKeybase() error {
	// Get private keys from manifest file
	_, current, _, _ := runtime.Caller(0)
	//nolint:gocritic // Use path to find private-keys yaml file from being called in any location in the repo
	yamlFile := filepath.Join(current, "../../../../../build/localnet/manifests/private-keys.yaml")

	if exists, err := fileExists(yamlFile); !exists || err != nil {
		return fmt.Errorf("Unable to find YAML file: %s", yamlFile)
	}

	// Parse the YAML file and load into the yamlConfig struct
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		return err
	}

	var config yamlConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	// Create/Open the keybase at `$HOME/.pocket/keys`
	kb, err := keybase.NewKeybase(DebugKeybasePath)
	if err != nil {
		return err
	}
	db := kb.GetDB()

	// Add the keys if the keybase contains less than 999
	curAddr, _, err := kb.GetAll()
	if err != nil {
		return err
	}

	// Add validator addresses if not present
	if len(curAddr) < validatorNumber {
		// Use writebatch to speed up bulk insert
		wb := db.NewWriteBatch()
		for _, privHexString := range config.StringData {
			// Import the keys into the keybase with no passphrase or hint as these are for debug purposes
			keyPair, err := crypto.CreateNewKeyFromString(privHexString, "", "")
			if err != nil {
				return err
			}

			// Use key address as key in DB
			addrKey := keyPair.GetAddressBytes()

			// Encode KeyPair into []byte for value
			keypairBz, err := keyPair.Marshal()
			if err != nil {
				return err
			}
			if err := wb.Set(addrKey, keypairBz); err != nil {
				return err
			}
		}
		if err := wb.Flush(); err != nil {
			return err
		}
	}

	// Close DB connection
	if err := kb.Stop(); err != nil {
		return err
	}

	return nil
}

// Check file at the given path exists
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
