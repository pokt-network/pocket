package keybase

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pokt-network/pocket/shared/utils"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/shared/crypto"
)

const (
	prefetchSize = 15
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
	pathExists, err := utils.DirExists(path) // Creates path if it doesn't exist
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

// Return DB instance
// FOR DEBUG PURPOSES ONLY
func (keybase *badgerKeybase) GetBadgerDB() *badger.DB {
	return keybase.db
}

// Close the DB
func (keybase *badgerKeybase) Stop() error {
	return keybase.db.Close()
}

// Create a new key and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) Create(passphrase, hint string) (keyPair crypto.KeyPair, err error) {
	err = keybase.db.Update(func(tx *badger.Txn) (err error) {
		keyPair, err = crypto.CreateNewKey(passphrase, hint)
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
	if err != nil {
		return nil, err
	}

	return keyPair, nil
}

// Create a new KeyPair from the private key hex string and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) ImportFromString(privKeyHex, passphrase, hint string) (keyPair crypto.KeyPair, err error) {
	err = keybase.db.Update(func(tx *badger.Txn) (err error) {
		keyPair, err = crypto.CreateNewKeyFromString(privKeyHex, passphrase, hint)
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
	if err != nil {
		return nil, err
	}

	return keyPair, nil
}

// Create a new KeyPair from the private key JSON string and store the serialised KeyPair encoding in the DB
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) ImportFromJSON(jsonStr, passphrase string) (keyPair crypto.KeyPair, err error) {
	err = keybase.db.Update(func(tx *badger.Txn) (err error) {
		keyPair, err = crypto.ImportKeyFromJSON(jsonStr, passphrase)
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
	if err != nil {
		return nil, err
	}

	return keyPair, nil
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
