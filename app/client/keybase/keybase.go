package keybase

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
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

type KeybaseType int

const (
	KeybaseTypeFile KeybaseType = iota
	KeybaseTypeVault
)

type KeybaseOptions struct {
	// Path to the keybase file
	KeybasePath string

	// Address of the keybase server
	VaultAddr string

	// Token for the keybase server
	VaultToken string

	// Mount path for the keybase server
	VaultMountPath string
}

// NewKeybase creates a new keybase based on the type and customized with the options provided
func NewKeybase(keybaseType KeybaseType, opts *KeybaseOptions) (Keybase, error) {
	switch keybaseType {
	case KeybaseTypeFile:
		// Open the file-based keybase at the specified path
		if opts == nil || opts.KeybasePath == "" {
			return nil, errors.New("keybase path is required for file-based keybase")
		}
		return NewBadgerKeybase(opts.KeybasePath)
	case KeybaseTypeVault:
		// Open the vault-based keybase
		if opts == nil || opts.VaultAddr == "" || opts.VaultToken == "" || opts.VaultMountPath == "" {
			return nil, errors.New("vault address, Token, and Mount are required for vault-based keybase")
		}
		return NewVaultKeybase(vaultKeybaseConfig{
			Address: opts.VaultAddr,
			Token:   opts.VaultToken,
			Mount:   opts.VaultMountPath,
		})
	default:
		return nil, fmt.Errorf("invalid keybase type: %d", keybaseType)
	}
}
