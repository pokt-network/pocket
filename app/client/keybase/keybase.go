package keybase

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/app/client/keybase/hashicorp"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Keybase interface implements the CRUD operations for the keybase
type Keybase interface {
	// Debug
	GetBadgerDB() (*badger.DB, error)

	// Close the DB connection
	Stop() error

	// Create new keypair entry in DB
	Create(passphrase, hint string) (crypto.KeyPair, error)
	// Insert a new keypair from the private key hex string provided into the DB
	ImportFromString(privStr, passphrase, hint string) (crypto.KeyPair, error)
	// Insert a new keypair from the JSON string of the encrypted private key into the DB
	ImportFromJSON(jsonStr, passphrase string) (crypto.KeyPair, error)

	// SLIPS-0010 Key Derivation
	// Deterministically generate, store and return the derived child key
	DeriveChildFromKey(masterAddrHex, passphrase string, childIndex uint32, childPassphrase, childHint string, shouldStore bool) (crypto.KeyPair, error)
	DeriveChildFromSeed(seed []byte, childIndex uint32, childPassphrase, childHint string, shouldStore bool) (crypto.KeyPair, error)

	// Accessors
	Get(address string) (crypto.KeyPair, error)
	GetPubKey(address string) (crypto.PublicKey, error)
	GetPrivKey(address, passphrase string) (crypto.PrivateKey, error)
	GetAll() (addresses []string, keyPairs []crypto.KeyPair, err error)

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

// NewKeybase creates a new keybase based on the type and customized with the options provided
func NewKeybase(conf *configs.KeybaseConfig) (Keybase, error) {
	switch conf.Type {
	case types.KeybaseType_FILE:
		// Open the file-based keybase at the specified path
		if conf == nil || conf.FilePath == "" {
			return nil, errors.New("keybase path is required for file-based keybase")
		}
		return NewBadgerKeybase(conf.FilePath)
	case types.KeybaseType_VAULT:
		return hashicorp.NewVaultKeybase(conf)
	default:
		return nil, fmt.Errorf("invalid keybase type: %d", conf.Type)
	}
}
