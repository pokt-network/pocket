package keybase

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
	"golang.org/x/crypto/scrypt"
)

const (
	// Scrypt params
	n    = 32768
	r    = 8
	p    = 1
	klen = 32
	// Sec param
	secParam = 12
	// No passphrase
	defaultPassphrase = "passphrase"
)

var (
	// Errors
	ErrorWrongPassphrase = errors.New("Can't decrypt private key: wrong passphrase")
)

func ErrorAddrNotFound(addr string) error {
	return fmt.Errorf("No key found with address: %s", addr)
}

func init() {
	gob.Register(poktCrypto.Ed25519PublicKey{})
	gob.Register(ed25519.PublicKey{})
	gob.Register(KeyPair{})
}

// KeyPair struct stores the public key and the passphrase encrypted private key
type KeyPair struct {
	PublicKey     poktCrypto.PublicKey
	PrivKeyArmour string
}

// Generate a new KeyPair struct given the public key and armoured private key
func NewKeyPair(pub poktCrypto.PublicKey, priv string) KeyPair {
	return KeyPair{
		PublicKey:     pub,
		PrivKeyArmour: priv,
	}
}

// Return the byte slice address of the public key
func (kp KeyPair) GetAddressBytes() []byte {
	return kp.PublicKey.Address().Bytes()
}

// Return the string address of the public key
func (kp KeyPair) GetAddressString() string {
	return kp.PublicKey.Address().String()
}

// Unarmour the private key with the passphrase provided
func (kp KeyPair) Unarmour(passphrase string) (poktCrypto.PrivateKey, error) {
	return unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
}

// Export Private Key String
func (kp KeyPair) ExportString(passphrase string) (string, error) {
	privKey, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return privKey.String(), nil
}

// Export Private Key as armoured JSON string with fields to decrypt
func (kp KeyPair) ExportJSON(passphrase string) (string, error) {
	_, err := unarmourDecryptPrivKey(kp.PrivKeyArmour, passphrase)
	if err != nil {
		return "", err
	}
	return kp.PrivKeyArmour, nil
}

// Armoured Private Key struct with fields to unarmour it later
type ArmouredKey struct {
	Kdf        string `json:"kdf"`
	Salt       string `json:"salt"`
	SecParam   string `json:"secparam"`
	Hint       string `json:"hint"`
	CipherText string `json:"ciphertext"`
}

// Generate new armoured private key struct with parameters for unarmouring
func NewArmouredKey(kdf, salt, hint, cipher string) ArmouredKey {
	return ArmouredKey{
		Kdf:        kdf,
		Salt:       salt,
		SecParam:   strconv.Itoa(secParam),
		Hint:       hint,
		CipherText: cipher,
	}
}

// Generate new private ED25519 key and encrypt and armour it as a string
// Returns a KeyPair struct of the Public Key and Armoured String
func CreateNewKey(passphrase string) (KeyPair, error) {
	privKey, err := poktCrypto.GeneratePrivateKey()
	if err != nil {
		return KeyPair{}, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase)
	if err != nil || privArmour == "" {
		return KeyPair{}, nil
	}

	pubKey := privKey.PublicKey()
	keyPair := NewKeyPair(pubKey, privArmour)

	return keyPair, nil
}

// Generate new KeyPair from the hex string provided, encrypt and armour it as a string
func CreateNewKeyFromString(privStr, passphrase string) (KeyPair, error) {
	privKey, err := poktCrypto.NewPrivateKey(privStr)
	if err != nil {
		return KeyPair{}, err
	}

	privArmour, err := encryptArmourPrivKey(privKey, passphrase)
	if err != nil || privArmour == "" {
		return KeyPair{}, nil
	}

	pubKey := privKey.PublicKey()
	keyPair := NewKeyPair(pubKey, privArmour)

	return keyPair, nil
}

// Create new KeyPair from the JSON encoded privStr
func ImportKeyFromJSON(jsonStr, passphrase string) (KeyPair, error) {
	// Get Private Key from armouredStr
	privKey, err := unarmourDecryptPrivKey(jsonStr, passphrase)
	if err != nil {
		return KeyPair{}, err
	}
	pubKey := privKey.PublicKey()
	keyPair := NewKeyPair(pubKey, jsonStr)

	return keyPair, nil
}

// Encrypt the given privKey with the passphrase, armour it by encoding the ecnrypted
// []byte into base64, and convert into a json string with the parameters for unarmouring
func encryptArmourPrivKey(privKey poktCrypto.PrivateKey, passphrase string) (string, error) {
	// Encrypt privKey usign AES-256 GCM Cipher
	saltBytes, encBytes, err := encryptPrivKey(privKey, passphrase)
	if err != nil {
		return "", err
	}

	// Armour encrypted bytes by encoding into Base64
	armourStr := base64.StdEncoding.EncodeToString(encBytes)

	// Create ArmouredKey object so can unarmour later
	armoured := NewArmouredKey("scrypt", fmt.Sprintf("%X", saltBytes), "", armourStr)

	// Encode armoured struct into []byte
	js, err := json.Marshal(armoured)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// Encrypt the given privKey with the passphrase using a randomly
// generated salt and the AES-256 GCM cipher
func encryptPrivKey(privKey poktCrypto.PrivateKey, passphrase string) (saltBytes []byte, encBytes []byte, err error) {
	// Use a default passphrase when none is given
	if passphrase == "" {
		passphrase = defaultPassphrase
	}

	// Get random bytes for salt
	saltBytes = randBytes(16)

	// Derive key for encryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	key, err := scrypt.Key([]byte(passphrase), saltBytes, n, r, p, klen)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt using AES
	privKeyBytes := privKey.Bytes()
	encBytes, err = encryptAESGCM(key, privKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	return saltBytes, encBytes, nil
}

// Unarmor and decrypt the private key using the passphrase provided
func unarmourDecryptPrivKey(armourStr string, passphrase string) (privKey poktCrypto.PrivateKey, err error) {
	// Decode armourStr back into ArmouredKey struct
	armouredKey := ArmouredKey{}
	err = json.Unmarshal([]byte(armourStr), &armouredKey)
	if err != nil {
		return nil, err
	}

	// Check the ArmouredKey for the correct parameters on kdf and salt
	if armouredKey.Kdf != "scrypt" {
		return nil, fmt.Errorf("Unrecognized KDF type: %v", armouredKey.Kdf)
	}
	if armouredKey.Salt == "" {
		return nil, fmt.Errorf("Missing salt bytes")
	}

	// Decoding the salt
	saltBytes, err := hex.DecodeString(armouredKey.Salt)
	if err != nil {
		return nil, fmt.Errorf("Error decoding salt: %v", err.Error())
	}

	// Decoding the "armoured" ciphertext stored in base64
	encBytes, err := base64.StdEncoding.DecodeString(armouredKey.CipherText)
	if err != nil {
		return nil, fmt.Errorf("Error decoding ciphertext: %v", err.Error())
	}

	// Decrypt the actual privkey with the parameters
	privKey, err = decryptPrivKey(saltBytes, encBytes, passphrase)

	return privKey, err
}

// Decrypt the AES-256 GCM encrypted bytes using the passphrase given
func decryptPrivKey(saltBytes []byte, encBytes []byte, passphrase string) (poktCrypto.PrivateKey, error) {
	// Use a default passphrase when none is given
	if passphrase == "" {
		passphrase = defaultPassphrase
	}

	// Derive key for decryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	key, err := scrypt.Key([]byte(passphrase), saltBytes, n, r, p, klen)
	if err != nil {
		return nil, err
	}

	// Decrypt using AES
	privKeyBytes, err := decryptAESGCM(key, encBytes)
	if err != nil {
		return nil, err
	}

	// Get private key from decrypted bytes
	pk, err := poktCrypto.NewPrivateKeyFromBytes(privKeyBytes)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// Encrypt using AES-256 GCM Cipher
func encryptAESGCM(key []byte, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := key[:12]
	out := gcm.Seal(nil, nonce, src, nil)
	return out, nil
}

// Decrypt using AES-256 GCM Cipher
func decryptAESGCM(key []byte, enBytes []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := key[:12]
	result, err := gcm.Open(nil, nonce, enBytes, nil)
	if err != nil && strings.Contains(err.Error(), "authentication failed") {
		return nil, ErrorWrongPassphrase
	} else if err != nil {
		return nil, fmt.Errorf("Can't Decrypt Using AES : %s \n", err)
	}
	return result, nil
}

// Use OS randomness
func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
