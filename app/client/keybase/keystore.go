package keybase

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Errors
func ErrorAddrNotFound(addr string) error {
	return fmt.Errorf("No key found with address: %s", addr)
}

// badgerKeybase implements the KeyBase interface
var _ Keybase = &badgerKeybase{}

// badgerKeybase implements the Keybase struct using the BadgerDB backend
type badgerKeybase struct {
	db *badger.DB
}

// Creates/Opens the DB at the specified path
func NewKeybase(path string) (Keybase, error) {
	pathExists, err := dirExists(path) // Creates path if it doesn't exist
	if err != nil || !pathExists {
		return nil, err
	}
	db, err := badger.Open(badgerOptions(path))
	if err != nil {
		return nil, err
	}
	return &badgerKeybase{db: db}, nil
}

// Creates/Opens the DB in Memory
// FOR TESTING PURPOSES ONLY
func NewKeybaseInMemory() (Keybase, error) {
	db, err := badger.Open(badgerOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}
	return &badgerKeybase{db: db}, nil
}

// Close the DB
func (keybase *badgerKeybase) Stop() error {
	return keybase.db.Close()
}

// Create a new key and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) Create(passphrase, hint string) error {
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
func (keybase *badgerKeybase) ImportFromString(privKeyHex, passphrase, hint string) error {
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
func (keybase *badgerKeybase) ImportFromJSON(jsonStr, passphrase string) error {
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
func (keybase *badgerKeybase) Get(address string) (crypto.KeyPair, error) {
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
func (keybase *badgerKeybase) GetPubKey(address string) (crypto.PublicKey, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return nil, err
	}

	return kp.GetPublicKey(), nil
}

// Returns a PrivateKey interface provided the address was found in the DB and the passphrase was correct
func (keybase *badgerKeybase) GetPrivKey(address, passphrase string) (crypto.PrivateKey, error) {
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
func (keybase *badgerKeybase) GetAll() (addresses []string, keyPairs []crypto.KeyPair, err error) {
	// View executes the function provided managing a read only transaction
	err = keybase.db.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 5
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

// Deterministically generate and return the ith child key from the masterAddrHex key stored in the keybase
func (keybase *badgerKeybase) DeriveChildFromKey(masterAddrHex, passphrase string, childIndex int32) (crypto.KeyPair, error) {
	privKey, err := keybase.GetPrivKey(masterAddrHex, passphrase)
	if err != nil {
		return nil, err
	}
	seed := privKey.Seed()
	path := fmt.Sprintf(crypto.PoktAccountPathFormat, childIndex)
	childKey, err := crypto.DeriveKeyFromPath(path, seed)
	if err != nil {
		return nil, err
	}
	return childKey, nil
}

// Deterministically generate and return the ith child from the seed provided
func (keybase *badgerKeybase) DeriveChildFromSeed(seed []byte, childIndex int32) (crypto.KeyPair, error) {
	path := fmt.Sprintf(crypto.PoktAccountPathFormat, childIndex)
	childKey, err := crypto.DeriveKeyFromPath(path, seed)
	if err != nil {
		return nil, err
	}
	return childKey, nil
}

// Deterministically generate and store the ith child from the masterAddrHex key stored in the keybase
func (keybase *badgerKeybase) StoreChildFromKey(masterAddrHex, masterPassphrase string, childIndex int32, childPassphrase, childHint string) error {
	masterPrivKey, err := keybase.GetPrivKey(masterAddrHex, masterPassphrase)
	if err != nil {
		return err
	}
	seed := masterPrivKey.Seed()
	path := fmt.Sprintf(crypto.PoktAccountPathFormat, childIndex)
	childKey, err := crypto.DeriveKeyFromPath(path, seed)
	if err != nil {
		return err
	}
	// No need to re-encrypt with provided passphrase
	if childPassphrase == "" && childHint == "" {
		err = keybase.db.Update(func(tx *badger.Txn) error {
			// Use key address as key in DB
			addrKey := childKey.GetAddressBytes()

			// Encode KeyPair into []byte for value
			keypairBz, err := childKey.Marshal()
			if err != nil {
				return err
			}

			return tx.Set(addrKey, keypairBz)
		})
	} else {
		// Re-encrypt child key with passphrase and hint
		err = keybase.db.Update(func(tx *badger.Txn) error {
			// Get the private key hex string from the child key
			privKeyHex, err := childKey.ExportString("") // No passphrase by default
			if err != nil {
				return err
			}

			keyPair, err := crypto.CreateNewKeyFromString(privKeyHex, childPassphrase, childHint)
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
	}
	return nil
}

// Deterministically generate and store the ith child from the masterAddrHex key stored in the keybase
func (keybase *badgerKeybase) StoreChildFromSeed(seed []byte, childIndex int32, childPassphrase, childHint string) error {
	path := fmt.Sprintf(crypto.PoktAccountPathFormat, childIndex)
	childKey, err := crypto.DeriveKeyFromPath(path, seed)
	if err != nil {
		return err
	}
	// No need to re-encrypt with provided passphrase
	if childPassphrase == "" && childHint == "" {
		err = keybase.db.Update(func(tx *badger.Txn) error {
			// Use key address as key in DB
			addrKey := childKey.GetAddressBytes()

			// Encode KeyPair into []byte for value
			keypairBz, err := childKey.Marshal()
			if err != nil {
				return err
			}

			return tx.Set(addrKey, keypairBz)
		})
	} else {
		// Re-encrypt child key with passphrase and hint
		err = keybase.db.Update(func(tx *badger.Txn) error {
			// Get the private key hex string from the child key
			privKeyHex, err := childKey.ExportString("") // No passphrase by default
			if err != nil {
				return err
			}

			keyPair, err := crypto.CreateNewKeyFromString(privKeyHex, childPassphrase, childHint)
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
	}
	return nil
}

// Check whether an address is currently stored in the DB
func (keybase *badgerKeybase) Exists(address string) (bool, error) {
	val, err := keybase.Get(address)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

// Export the Private Key string of the given address
func (keybase *badgerKeybase) ExportPrivString(address, passphrase string) (string, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return "", err
	}
	return kp.ExportString(passphrase)
}

// Export the Private Key of the given address as a JSON object
func (keybase *badgerKeybase) ExportPrivJSON(address, passphrase string) (string, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return "", err
	}
	return kp.ExportJSON(passphrase)
}

func (keybase *badgerKeybase) UpdatePassphrase(address, oldPassphrase, newPassphrase, hint string) error {
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
func (keybase *badgerKeybase) Sign(address, passphrase string, msg []byte) ([]byte, error) {
	privKey, err := keybase.GetPrivKey(address, passphrase)
	if err != nil {
		return nil, err
	}
	return privKey.Sign(msg)
}

// Verify a message has been signed correctly
func (keybase *badgerKeybase) Verify(address string, msg, sig []byte) (bool, error) {
	kp, err := keybase.Get(address)
	if err != nil {
		return false, err
	}
	pubKey := kp.GetPublicKey()
	return pubKey.Verify(msg, sig), nil
}

// Remove a KeyPair from the DB given the address
func (keybase *badgerKeybase) Delete(address, passphrase string) error {
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
	if os.IsNotExist(err) {
		// Create directories in path recursively
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return false, fmt.Errorf("Error creating directory at path: %s, (%v)", path, err.Error())
		}
		return true, nil
	}
	return false, err
}
