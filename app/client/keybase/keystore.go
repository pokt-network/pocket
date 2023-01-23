package keybase

import (
	"bytes"
	"crypto/ed25519"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/shared/crypto"
)

func init() {
	gob.Register(crypto.Ed25519PublicKey{})
	gob.Register(ed25519.PublicKey{})
	gob.Register(KeyPair{})
	gob.Register(ArmouredKey{})
}

// badgerKeybase implements the KeyBase interface
var _ Keybase = &badgerKeybase{}

// Keybase interface implements the CRUD operations for the keybase
type Keybase interface {
	// Close the DB connection
	Stop() error

	// Create new keypair entry in DB
	Create(passphrase string) error
	// Insert a new keypair from the private key hex string provided into the DB
	ImportFromString(privStr, passphrase string) error

	// Accessors
	Get(address string) (KeyPair, error)
	//GetPubKey(address string) (crypto.PublicKey, error)
	GetPrivKey(address, passphrase string) (crypto.PrivateKey, error)
	GetAll() (addresses [][]byte, keyPairs []KeyPair, err error)
	Exists(address string) (bool, error)

	// Updator
	UpdatePassphrase(address, oldPassphrase, newPassphrase string) error

	// Sign Messages
	Sign(address, passphrase string, msg []byte) ([]byte, error)
	Verify(address string, msg, sig []byte) (bool, error)

	// Removals
	Delete(address, passphrase string) error
}

// badgerKeybase implements the Keybase struct using the BadgerDB backend
type badgerKeybase struct {
	db *badger.DB
}

// Creates/Opens the DB at the specified path
// WARNING: path must be a valid directory that already exists
func NewKeybase(path string) (Keybase, error) {
	db, err := badger.Open(badgerOptions(path))
	if err != nil {
		return nil, err
	}
	return &badgerKeybase{db: db}, nil
}

// Creates/Opens the DB in Memory
// FOR TESTING PURPOSES ONLY
func NewKeybaseInMemory(path string) (Keybase, error) {
	db, err := badger.Open(badgerOptions(path).WithInMemory(true))
	if err != nil {
		return nil, err
	}
	return &badgerKeybase{db: db}, nil
}

// Close the DB
func (keybase *badgerKeybase) Stop() error {
	return keybase.db.Close()
}

// Crate a new key and store it in the DB by encoding the KeyPair struct into a []byte
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) Create(passphrase string) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := CreateNewKey(passphrase)
		if err != nil {
			return err
		}

		// Use key address as key in DB
		key := keyPair.GetAddressBytes()
		// Encode entire KeyPair struct into []byte for value
		bz := new(bytes.Buffer)
		enc := gob.NewEncoder(bz)
		if err = enc.Encode(keyPair); err != nil {
			return err
		}

		err = tx.Set(key, bz.Bytes())
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// Crate a new KeyPair from the private key hex string and store it in the DB by encoding the KeyPair struct into a []byte
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) ImportFromString(privStr, passphrase string) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := CreateNewKeyFromString(privStr, passphrase)
		if err != nil {
			return err
		}

		// Use key address as key in DB
		key := keyPair.GetAddressBytes()
		// Encode entire KeyPair struct into []byte for value
		bz := new(bytes.Buffer)
		enc := gob.NewEncoder(bz)
		if err = enc.Encode(keyPair); err != nil {
			return err
		}

		err = tx.Set(key, bz.Bytes())
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// Returns a KeyPair struct provided the address was found in the DB
func (keybase *badgerKeybase) Get(address string) (KeyPair, error) {
	var kp KeyPair
	bz := new(bytes.Buffer)
	addrBz, err := hex.DecodeString(address)
	if err != nil {
		return KeyPair{}, err
	}

	err = keybase.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(addrBz)
		if err != nil && strings.Contains(err.Error(), "Key not found") {
			return ErrorAddrNotFound(address)
		} else if err != nil {
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// Decode []byte value back into KeyPair struct
		bz.Write(value)
		dec := gob.NewDecoder(bz)
		if err = dec.Decode(&kp); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return KeyPair{}, err
	}

	return kp, nil
}

// Returns a PrivateKey interface provided the address was found in the DB and the passphrase was correct
func (keybase *badgerKeybase) GetPrivKey(address, passphrase string) (crypto.PrivateKey, error) {
	var kp KeyPair
	bz := new(bytes.Buffer)
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

		// Decode []byte value back into KeyPair struct
		bz.Write(value)
		dec := gob.NewDecoder(bz)
		if err = dec.Decode(&kp); err != nil {
			return err
		}

		return nil
	})
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
func (keybase *badgerKeybase) GetAll() (addresses [][]byte, keyPairs []KeyPair, err error) {
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
				// Decode []byte value back into KeyPair struct
				var kp KeyPair
				bz := new(bytes.Buffer)
				bz.Write(b)
				dec := gob.NewDecoder(bz)
				if err = dec.Decode(&kp); err != nil {
					return err
				}

				addresses = append(addresses, item.Key())
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
func (keybase *badgerKeybase) Exists(address string) (bool, error) {
	val, err := keybase.Get(address)
	if err != nil {
		return false, err
	}
	return val != KeyPair{}, nil
}

func (keybase *badgerKeybase) UpdatePassphrase(address, oldPassphrase, newPassphrase string) error {
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
		keyPair, err := CreateNewKeyFromString(privStr, newPassphrase)
		if err != nil {
			return err
		}

		// Use key address as key in DB
		key := keyPair.GetAddressBytes()
		if bytes.Compare(key, addrBz) != 0 {
			return fmt.Errorf("Key address does not match previous address.")
		}
		// Encode entire KeyPair struct into []byte for value
		bz := new(bytes.Buffer)
		enc := gob.NewEncoder(bz)
		if err = enc.Encode(keyPair); err != nil {
			return err
		}

		err = tx.Set(key, bz.Bytes())
		if err != nil {
			return err
		}

		return nil
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
	pubKey := kp.PublicKey
	return pubKey.Verify(msg, sig), nil
}

// Remove a KeyPair from the DB given the address
func (keybase *badgerKeybase) Delete(address, passphrase string) error {
	_, err := keybase.GetPrivKey(address, passphrase)
	if err != nil {
		return err
	}

	addrBz, err := hex.DecodeString(address)
	if err != nil {
		return err
	}

	err = keybase.db.Update(func(tx *badger.Txn) error {
		tx.Delete(addrBz)
		return nil
	})
	return err
}

// Return badger.Options for the given DB path - disable logging
func badgerOptions(path string) badger.Options {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Badger logger is very noisy
	return opts
}
