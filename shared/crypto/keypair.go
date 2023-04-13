package crypto

import (
	"bytes"
	"crypto/ed25519"
	"encoding/gob"
	"encoding/json"
)

// Gob Encoding is used to serialise the data to store the KeyPairs in the BadgerDB database
func init() {
	gob.Register(Ed25519PublicKey{})
	gob.Register(ed25519.PublicKey{})
	gob.Register(encKeyPair{})
}

// KeyPair interface exposes functions relating to public and encrypted private key pairs
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

	// Seed
	GetSeed(passphrase string) ([]byte, error)

	// Marshalling
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

var _ KeyPair = &encKeyPair{}

// encKeyPair struct stores the public key and the passphrase encrypted private key
// The encrypted private key is stored as a JSON string with the fields needed to decrypt the key
type encKeyPair struct {
	PublicKey     PublicKey `json:"public_key"`
	PrivKeyArmour string    `json:"priv_key_armour"`
}

// newKeyPair Generate a new KeyPair struct given the public key and armoured private key
func newKeyPair(pub PublicKey, priv string) KeyPair {
	return &encKeyPair{
		PublicKey:     pub,
		PrivKeyArmour: priv,
	}
}

// GetKeypair Returns an empty KeyPair interface
func GetKeypair() KeyPair {
	return &encKeyPair{}
}

// GetPublicKey Returns the public key
func (kp encKeyPair) GetPublicKey() PublicKey {
	return kp.PublicKey
}

// GetPrivArmour Returns private key armoured string
func (kp encKeyPair) GetPrivArmour() string {
	return kp.PrivKeyArmour
}

// GetAddressBytes Returns the byte slice address of the public key
func (kp encKeyPair) GetAddressBytes() []byte {
	return kp.PublicKey.Address().Bytes()
}

// GetAddressString Returns the string address of the public key
func (kp encKeyPair) GetAddressString() string {
	return kp.PublicKey.Address().String()
}

// Unarmour the private key with the passphrase provided
func (kp encKeyPair) Unarmour(passphrase string) (PrivateKey, error) {
	return unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
}

// ExportString Exports the private key as a raw string
func (kp encKeyPair) ExportString(passphrase string) (string, error) {
	privKey, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return privKey.String(), nil
}

// ExportJSON Exports the private key as an armoured JSON string with fields to decrypt
func (kp encKeyPair) ExportJSON(passphrase string) (string, error) {
	_, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return kp.PrivKeyArmour, nil
}

// GetSeed Returns the seed of the key
func (kp encKeyPair) GetSeed(passphrase string) ([]byte, error) {
	privKey, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return []byte{}, err
	}
	return privKey.Seed(), nil
}

// Marshal Marshals the KeyPair into a []byte
func (kp encKeyPair) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(kp); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal Unmarshals a []byte into an encKeyPair struct
func (kp *encKeyPair) Unmarshal(bz []byte) error {
	var keyPair encKeyPair
	keypairBz := new(bytes.Buffer)
	keypairBz.Write(bz)
	dec := gob.NewDecoder(keypairBz)
	if err := dec.Decode(&keyPair); err != nil {
		return err
	}
	*kp = keyPair
	return nil
}

// UnmarshalJSON Unmarshals a JSON string into an encKeyPair struct
func (ekp *encKeyPair) UnmarshalJSON(data []byte) error {
	type Alias encKeyPair
	aux := &struct {
		PublicKey string `json:"public_key"`
		*Alias
	}{
		Alias: (*Alias)(ekp),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	pubKey, err := NewPublicKey(aux.PublicKey)
	if err != nil {
		return err
	}
	ekp.PublicKey = pubKey
	return nil
}

// CreateNewKey Generates a new private ED25519 key and encrypt and armour it as a string
// Returns a KeyPair interface and error if any
func CreateNewKey(passphrase, hint string) (KeyPair, error) {
	privKey, err := GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase, hint)
	if err != nil || privArmour == "" {
		return nil, err
	}

	pubKey := privKey.PublicKey()
	kp := newKeyPair(pubKey, privArmour)

	return kp, nil
}

// CreateNewKeyFromString Generates new KeyPair from the hex string provided, encrypt and armour it as a string
// Returns a KeyPair interface and error if any
func CreateNewKeyFromString(privStrHex, passphrase, hint string) (KeyPair, error) {
	privKey, err := NewPrivateKey(privStrHex)
	if err != nil {
		return nil, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase, hint)
	if err != nil || privArmour == "" {
		return nil, err
	}

	pubKey := privKey.PublicKey()
	kp := newKeyPair(pubKey, privArmour)

	return kp, nil
}

// CreateNewKeyFromSeed Generates a new KeyPair from the seed provided
// Returns a KeyPair interface and error if any
func CreateNewKeyFromSeed(seed []byte, passphrase, hint string) (KeyPair, error) {
	// Generate PrivateKey interface form secret key
	reader := bytes.NewReader(seed)
	privKey, err := GeneratePrivateKeyWithReader(reader)
	if err != nil {
		return nil, err
	}

	// Armour and encrypt private key into JSON string
	armouredStr, err := encryptArmourPrivKey(privKey, passphrase, hint) // No passphrase or hint as they depend on the master key
	if err != nil {
		return nil, err
	}

	// Return KeyPair interface
	return &encKeyPair{
		PublicKey:     privKey.PublicKey(),
		PrivKeyArmour: armouredStr,
	}, nil
}

// ImportKeyFromJSON Creates new KeyPair from the JSON encoded privStr
// Returns a KeyPair interface and error if any
func ImportKeyFromJSON(jsonStr, passphrase string) (KeyPair, error) {
	// Get Private Key from armouredStr
	privKey, err := unarmourDecryptPrivKey(jsonStr, passphrase)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.PublicKey()
	kp := newKeyPair(pubKey, jsonStr)

	return kp, nil
}
