package keybase

import (
	"bytes"
	"crypto/ed25519"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/shared/crypto"
)

func init() {
	gob.Register(crypto.Ed25519PublicKey{})
	gob.Register(ed25519.PublicKey{})
	gob.Register(KeyPair{})
	gob.Register(ArmouredKey{})
}

// Keybase interface implements the CRUD operations for the keybase
type Keybase interface {
	// Close the DB connection
	Stop() error

	// Create new keypair entry in DB
	Create(passphrase string) error
	// Insert new keypair from private key []byte provided into the DB
	CreateFromBytes(privBytes []byte, passphrase string) error

	// Accessors
	Get(address []byte) (KeyPair, error)
	GetAll() (addresses [][]byte, keyPairs []KeyPair, err error)
	Exists(key []byte) (bool, error)

	// Removals
	Delete(address []byte) error
	ClearAll() error
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
		key := keyPair.GetAddress()
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

// Crate a new key and store it in the DB by encoding the KeyPair struct into a []byte
// Using the PublicKey.Address() return value as the key for storage
func (keybase *badgerKeybase) CreateFromBytes(privBytes []byte, passphrase string) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		keyPair, err := CreateNewKeyFromBytes(privBytes, passphrase)
		if err != nil {
			return err
		}

		// Use key address as key in DB
		key := keyPair.GetAddress()
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
func (keybase *badgerKeybase) Get(address []byte) (KeyPair, error) {
	var kp KeyPair
	bz := new(bytes.Buffer)

	err := keybase.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(address)
		if err != nil {
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
				// Decode []byte value back into KeyPair struct
				var kp KeyPair
				bz := new(bytes.Buffer)
				bz.Write(val)
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
func (keybase *badgerKeybase) Exists(address []byte) (bool, error) {
	val, err := keybase.Get(address)
	if err != nil {
		return false, err
	}
	return val == KeyPair{}, nil
}

// Remove a KeyPair from the DB given the address
// TODO: Add a check that the user can decrypt this KeyPair
func (keybase *badgerKeybase) Delete(address []byte) error {
	err := keybase.db.Update(func(tx *badger.Txn) error {
		tx.Delete(address)
		return nil
	})
	return err
}

// Remove all keys in the DB
// TODO: Add a check that the use can decrypt all the keys
func (keybase *badgerKeybase) ClearAll() error {
	return keybase.db.DropAll()
}

// Return badger.Options for the given DB path - disable logging
func badgerOptions(path string) badger.Options {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Badger logger is very noisy
	return opts
}
