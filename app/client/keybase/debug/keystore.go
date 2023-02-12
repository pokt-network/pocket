package debug

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/shared/crypto"
	"gopkg.in/yaml.v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	prefetchSize       = 15
	validatorNumber    = 999
	debugKeybaseSuffix = "/.pocket/keys"
)

var (
	DebugKeybasePath string
)

// Errors
func ErrorAddrNotFound(addr string) error {
	return fmt.Errorf("No key found with address: %s", addr)
}

// Initialise the debug keybase with the 999 validator keys from the private-keys manifest file
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("[ERROR] Cannot find user home directory: %s", err.Error())
	}
	DebugKeybasePath = homeDir + debugKeybaseSuffix

	if err := InitialiseDebugKeybase(DebugKeybasePath); err != nil { // Initialise the debug keybase with the 999 validators
		log.Fatalf("[ERROR] Cannot initialise the keybase with the validator keys: %s", err.Error())
	}
}

// badgerKeybase implements the KeyBase interface
var _ keybase.Keybase = &debugBadgerKeybase{}

// badgerKeybase implements the Keybase struct using the BadgerDB backend
type debugBadgerKeybase struct {
	db *badger.DB
}

// Creates/Opens the DB at `$HOME/.pocket/keys`
func NewDebugKeybase(path string) (keybase.Keybase, error) {
	pathExists, err := dirExists(path) // Creates path if it doesn't exist
	if err != nil || !pathExists {
		return nil, err
	}
	db, err := badger.Open(badgerOptions(path))
	if err != nil {
		return nil, err
	}
	return &debugBadgerKeybase{db: db}, nil
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
func InitialiseDebugKeybase(path string) error {
	// Get private keys from manifest file
	_, current, _, _ := runtime.Caller(0)
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

	// Create/Open the keybase
	pathExists, err := dirExists(path) // Creates path if it doesn't exist
	if err != nil || !pathExists {
		return err
	}
	db, err := badger.Open(badgerOptions(path))
	if err != nil {
		return err
	}
	kb := &debugBadgerKeybase{db: db}

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

// Close the DB
func (keybase *debugBadgerKeybase) Stop() error {
	return keybase.db.Close()
}

// Create a new key and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *debugBadgerKeybase) Create(passphrase, hint string) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := crypto.CreateNewKey(passphrase, hint)
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

		return tx.Set(addrKey, keypairBz)
	})

	return err
}

// Create a new KeyPair from the private key hex string and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *debugBadgerKeybase) ImportFromString(privKeyHex, passphrase, hint string) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := crypto.CreateNewKeyFromString(privKeyHex, passphrase, hint)
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

		return tx.Set(addrKey, keypairBz)
	})

	return err
}

// Create a new KeyPair from the private key JSON string and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *debugBadgerKeybase) ImportFromJSON(jsonStr, passphrase string) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := crypto.ImportKeyFromJSON(jsonStr, passphrase)
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

		return tx.Set(addrKey, keypairBz)
	})

	return err
}

// Returns a KeyPair struct provided the address was found in the DB
func (keybase *debugBadgerKeybase) Get(address string) (crypto.KeyPair, error) {
	kp := crypto.GetKeypair()
	addrBz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	err = keybase.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(addrBz)
		if err != nil && strings.Contains(err.Error(), "not found") {
			return ErrorAddrNotFound(address)
		} else if err != nil {
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// Decode []byte value back into KeyPair
		if err := kp.Unmarshal(value); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return kp, nil
}

// Returns a PublicKey interface provided the address was found in the DB
func (keybase *debugBadgerKeybase) GetPubKey(address string) (crypto.PublicKey, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return nil, err
	}

	return kp.GetPublicKey(), nil
}

// Returns a PrivateKey interface provided the address was found in the DB and the passphrase was correct
func (keybase *debugBadgerKeybase) GetPrivKey(address, passphrase string) (crypto.PrivateKey, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return nil, err
	}

	privKey, err := kp.Unarmour(passphrase)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// Get all the addresses and key pairs stored in the keybase
// Returns addresses stored and all the KeyPair structs stored in the DB
func (keybase *debugBadgerKeybase) GetAll() (addresses []string, keyPairs []crypto.KeyPair, err error) {
	// View executes the function provided managing a read only transaction
	err = keybase.db.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				b := make([]byte, len(val))
				copy(b, val)

				// Decode []byte value back into KeyPair
				kp := crypto.GetKeypair()
				if err := kp.Unmarshal(b); err != nil {
					return err
				}

				addresses = append(addresses, kp.GetAddressString())
				keyPairs = append(keyPairs, kp)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return addresses, keyPairs, nil
}

// Check whether an address is currently stored in the DB
func (keybase *debugBadgerKeybase) Exists(address string) (bool, error) {
	val, err := keybase.Get(address)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

// Export the Private Key string of the given address
func (keybase *debugBadgerKeybase) ExportPrivString(address, passphrase string) (string, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return "", err
	}
	return kp.ExportString(passphrase)
}

// Export the Private Key of the given address as a JSON object
func (keybase *debugBadgerKeybase) ExportPrivJSON(address, passphrase string) (string, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return "", err
	}
	return kp.ExportJSON(passphrase)
}

func (keybase *debugBadgerKeybase) UpdatePassphrase(address, oldPassphrase, newPassphrase, hint string) error {
	// Check the oldPassphrase is correct
	privKey, err := keybase.GetPrivKey(address, oldPassphrase)
	if err != nil {
		return err
	}
	privStr := privKey.String()

	addrBz, err := hex.DecodeString(address)
	if err != nil {
		return err
	}

	err = keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := crypto.CreateNewKeyFromString(privStr, newPassphrase, hint)
		if err != nil {
			return err
		}

		// Use key address as key in DB
		addrKey := keyPair.GetAddressBytes()
		if !bytes.Equal(addrKey, addrBz) {
			return fmt.Errorf("Key address does not match previous address.")
		}

		// Encode KeyPair into []byte for value
		keypairBz, err := keyPair.Marshal()
		if err != nil {
			return err
		}

		return tx.Set(addrKey, keypairBz)
	})

	return err
}

// Sign a message using the key address provided
func (keybase *debugBadgerKeybase) Sign(address, passphrase string, msg []byte) ([]byte, error) {
	privKey, err := keybase.GetPrivKey(address, passphrase)
	if err != nil {
		return nil, err
	}
	return privKey.Sign(msg)
}

// Verify a message has been signed correctly
func (keybase *debugBadgerKeybase) Verify(address string, msg, sig []byte) (bool, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return false, err
	}
	pubKey := kp.GetPublicKey()
	return pubKey.Verify(msg, sig), nil
}

// Remove a KeyPair from the DB given the address
func (keybase *debugBadgerKeybase) Delete(address, passphrase string) error {
	if _, err := keybase.GetPrivKey(address, passphrase); err != nil {
		return err
	}

	addrBz, err := hex.DecodeString(address)
	if err != nil {
		return err
	}

	err = keybase.db.Update(func(tx *badger.Txn) error {
		return tx.Delete(addrBz)
	})
	return err
}

// Return badger.Options for the given DB path - disable logging
func badgerOptions(path string) badger.Options {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Badger logger is very noisy
	return opts
}

// Check directory exists and creates path if it doesn't exist
func dirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		// Exists but not directory
		if !stat.IsDir() {
			return false, fmt.Errorf("Keybase path is not a directory: %s", path)
		}
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		// Create directories in path recursively
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return false, fmt.Errorf("Error creating directory at path: %s, (%v)", path, err.Error())
		}
		return true, nil
	}
	return false, err
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
