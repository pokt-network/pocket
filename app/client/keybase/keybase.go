package keybase

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Keybase interface implements the CRUD operations for the keybase
type Keybase interface {
	// Debug
	GetBadgerDB() *badger.DB

	// Close the DB connection
	Stop() error

	// Create new keypair entry in DB
	Create(passphrase, hint string) error
	// Insert a new keypair from the private key hex string provided into the DB
	ImportFromString(privStr, passphrase, hint string) error
	// Insert a new keypair from the JSON string of the encrypted private key into the DB
	ImportFromJSON(jsonStr, passphrase string) error

	// SLIPS-0010 Key Derivation
	// Deterministically generate and return the derived child key
	DeriveChildFromKey(masterAddrHex, passphrase string, childIndex uint32) (crypto.KeyPair, error)
	DeriveChildFromSeed(seed []byte, childIndex uint32) (crypto.KeyPair, error)
	// Deterministically generate and store the derived child key in the keybase
	StoreChildFromKey(masterAddrHex, masterPassphrase string, childIndex uint32, childPassphrase, childHint string) error
	StoreChildFromSeed(seed []byte, childIndex uint32, childPassphrase, childHint string) error

	// Accessors
	Get(address string) (crypto.KeyPair, error)
	GetPubKey(address string) (crypto.PublicKey, error)
	GetPrivKey(address, passphrase string) (crypto.PrivateKey, error)
	GetAll() (addresses []string, keyPairs []crypto.KeyPair, err error)
	Exists(address string) (bool, error)

	// Exporters
	ExportPrivString(address, passphrase string) (string, error)
	ExportPrivJSON(address, passphrase string) (string, error)

	// Updator
	UpdatePassphrase(address, oldPassphrase, newPassphrase, hint string) error

	// Sign Messages
	Sign(address, passphrase string, msg []byte) ([]byte, error)
	Verify(address string, msg, sig []byte) (bool, error)

	// Removals
	Delete(address, passphrase string) error
}
