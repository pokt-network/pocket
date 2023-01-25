# Keybase

This document is intended to outline the current Keybase implementation used by the V1 client, and is primarily focused on its design and implementation as well as testing.

- [Backend Database](#backend-database)
- [Keybase Interface](#keybase-interface)
  - [Code structure](#keybase-code-structure)
  - [Makefile helper](#makefile-helper)
- [KeyPair Interface](#keypair-interface)
  - [Code structure](#keypair-code-structure)
- [Encryption and Armouring](#encryption-and-armouring)
- [Testing](#testing)
- [TODO](#todo)


## Backend Database

The Keybase uses the filesystem database `BadgerDB` as its backend to persistently store keys locally on the clients machine, in a key-value store pattern.


_The current keybase has not been integrated with any CLI endpoints, and as such is only accessible through the [keybase interface](#keybase-interface)_


The DB stores the local key pairs in `EncKeyPair` structs encoded into `[]byte` using `encoding/gob` this is only used for internal storage in the DB. The `EncKeyPair` struct implements the [KeyPair interface](#keypair-interface) and as such has a number of methods that can be used on it. But relevent to the DB storage of these is the `GetAddressBytes()` function that returns the `[]byte` of the `PublicKey` field's hex address from the struct. The `[]byte` returned by the `GetAddressBytes()` function is used as the key in the key-value store and the value is the `gob` encoded `[]byte` of the `EncKeyPair` struct as a whole - which contains both the `PublicKey` and `PrivKeyArmour` (JSON encoded, encrypted private key string).


The Keybase DB layer then allows for a number of functions to be used which are exposed by the [Keybase interface](#keybase-interface) to fulfill CRUD operations on the DB itself.


## Keybase Interface

The Keybase interface exposes the CRUD operations for interacting with the database layer.

```go
// Keybase interface implements the CRUD operations for the keybase
type Keybase interface {
	// Close the DB connection
	Stop() error

	// Create new keypair entry in DB
	Create(passphrase string) error
	// Insert a new keypair from the private key hex string provided into the DB
	ImportFromString(privStr, passphrase string) error
	// Insert a new keypair from the JSON string of the encrypted private key into the DB
	ImportFromJSON(jsonStr, passphrase string) error

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
	UpdatePassphrase(address, oldPassphrase, newPassphrase string) error

	// Sign Messages
	Sign(address, passphrase string, msg []byte) ([]byte, error)
	Verify(address string, msg, sig []byte) (bool, error)

	// Removals
	Delete(address, passphrase string) error
}
```

The `Keybase` interface allows for the import/export of keys between V0<->V1. Meaning any key created in the V0 protocol can be imported in two ways to the V1 protocol.
 1. Via the JSON keyfile
    - This method will take the JSON encoded, encrypted private key, and will import it into the V1 keybase - the `passphrase` supplied must be the same as the one use to encrypt the key in the first place or the key won't be able to be imported
 2. Via the private key hex string
    - This method will directly import the private key from the hex string provided and then encrypt it with the passphrase provided - this does mean than the passphrase can be different from the original as this is a decrypted form of the private key


Although key pairs are stored in the local DB using the `[]byte` of the public key address as the key for retrieval all the accessing methods use the hex string of the public key's address to actually find the key for ease of use.


Keys can be created without the use of any password - in order to do this the `passphrase` supplied to the functions must be `""`. The private key will still be encrypted but will simply use the empty string as the key.


### Keybase Code Structure
```
[ 90] app/client/keybase
├──── [ 8KB] README.md
├──── [ 1KB] keybase.go
├──── [ 13KB] keybase_test.go
└──── [ 8KB] keystore.go
```


The interface itself is found in `app/client/keybase/keybase.go` whereas its implementation can be found in `app/client/keybase/keystore.go`


### Makefile Helper


To aid in the testing of the local keybase the following `Makefile` command has been exposed `make test_app` which will run the test suites from the `app` module alone, which includes the `app/client/keybase/keybase_test.go` file which covers the functionality of the `Keybase` implementation


## KeyPair Interface


The `KeyPair` interface exposes methods related to the operations used on the pairs of `PublicKey` types and JSON encoded, `PrivKeyArmour` strings.

```go
// The KeyPair interface exposes functions relating to public and encrypted private key pairs
type KeyPair interface {
	// Accessors
	GetPublicKey() PublicKey
	GetPrivArmour() string

	// Public Key Address
	GetAddressBytes() []byte
	GetAddressString() string // hex string

	// Unarmour
	Unarmour(passphrase string) (PrivateKey, error)

	// Export
	ExportString(passphrase string) (string, error)
	ExportJSON(passphrase string) (string, error)
}
```

The `KeyPair` interface is implemented by the `EncKeyPair` struct

```go
// EncKeyPair struct stores the public key and the passphrase encrypted private key
type EncKeyPair struct {
	PublicKey     PublicKey
	PrivKeyArmour string
}
```

The `EncKeyPair` struct stores the `PublicKey` of the key pair and the JSON encoded, `"armoured"` key string. The armoured key string is built from the following struct

```go
// Armoured Private Key struct with fields to unarmour it later
type ArmouredKey struct {
	Kdf        string `json:"kdf"`
	Salt       string `json:"salt"`
	SecParam   string `json:"secparam"`
	Hint       string `json:"hint"`
	CipherText string `json:"ciphertext"`
}
```

This struct is created after the [encryption step](#encryption-and-armouring) has encrypted the private key string into the `CipherText` and the other fields are filled using the parameters used in the [encryption step](#encryption-and-armouring) so that the armoured key can later be unarmoured and decrypted provided the correct passphrase is supplied. This struct is then marshalled into a JSON string and stored in the `EncKeyPair`'s `PrivKeyArmour` field.


## KeyPair Code Structure

The KeyPair code is seperated into two files `shared/crypto/keypair.go` and `shared/crypto/armour.go`

```
[ 146] shared
└──── [ 102] crypto
      ├──── [ 5KB] armour.go
      └──── [ 3KB] keypair.go
```


## Encryption and Armouring

Whenever a new key is created or imported it is encrypted using the passphrase provided (this can be `""` for no passphrase).

The packages used for the encryption and armouring of a key are: `golang.org/x/crypto/scrypt` (key generation), `crypto/aes`, `crypto/cipher` (AES256-GCM encryption cipher) as well as `encoding/hex`, `encoding/base64` and `encoding/json` for the armouring of the encrypted key.

1. First the OS's randomness is used to generate a `[]byte` of length `16` which is used as the salt for generating the key.
2. `scrypt` is then used to generate a `[]byte` of length 32, which is used as the key for the `AES256-GCM` encryption cipher.
3. The `AES256-GCM` cipher then uses the first `12` bytes of the key to encrypt the `[]byte` conversion of the private key hex string.
   - _NOTE_ It is important that the cipher encrypts `[]byte(privateKeyHexString)` for interoperability with V0 keys.

The armouring process then encodes the encrypted `[]byte` using the `base64.RawStdEncoding.EncodeToString()` function into a hex string. This is the `CipherText` field of the `ArmouredKey` struct. Along with this the salt bytes and other parameters used to encrypt and encode this key are put into an `ArmouredKey` struct before being marshalled into a JSON string and stored in the `EncKeyPair.PrivKeyArmour` field of a KeyPair.


When unarmouring the same process is done in reverse - the `EncKeyPair.PrivKeyArmour` JSON string is unmarshalled into the `AmouredKey` struct and the process is done in reverse. After the decryption of the `CipherText` the `[]byte` returned must be converted to a string and decoded from its hex from to be turned back into a `PrivateKey` type. This again is for interoperability with V0.


## Testing

The full test suite can be run with `make test_app` where the [Keybase interface's](#keybase-interface) methods are tested with unit tests.


## TODO

- [ ] Add better error catching and error messages for importing keys with invalid strings/invalid JSON
- [ ] Research and implement threshold signatures and threshold keys
- [ ] Look into a fully feature signature implementation beyond trivial `[]byte` messages
