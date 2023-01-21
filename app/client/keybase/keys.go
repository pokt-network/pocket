package keybase

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
	"golang.org/x/crypto/scrypt"
)

const (
	// Scrypt params
	n    = 32768
	r    = 8
	p    = 1
	klen = 32
)

func init() {
	gob.Register(poktCrypto.Ed25519PublicKey{})
	gob.Register(ed25519.PublicKey{})
	gob.Register(KeyPair{})
	gob.Register(ArmouredKey{})
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
func (kp KeyPair) GetAddress() []byte {
	return kp.PublicKey.Address().Bytes()
}

// Armoured Private Key struct with fields to unarmour it later
type ArmouredKey struct {
	Kdf        string
	Salt       string
	CipherText string
}

// Generate new armoured private key struct with parameters for unarmouring
func NewArmouredKey(kdf, salt, cipher string) ArmouredKey {
	return ArmouredKey{
		Kdf:        kdf,
		Salt:       salt,
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

// Generate new private ED25519 key from the bytes provided, encrypt and
// armour it as a string
// Returns a KeyPair struct of the Public Key and Armoured String
func CreateNewKeyFromBytes(privBytes []byte, passphrase string) (KeyPair, error) {
	privKey, err := poktCrypto.NewPrivateKeyFromBytes(privBytes)
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
	armoured := NewArmouredKey("scrypt", fmt.Sprintf("%X", saltBytes), armourStr)

	// Encode armoured struct into []byte
	var bz bytes.Buffer
	enc := gob.NewEncoder(&bz)
	if err = enc.Encode(armoured); err != nil {
		return "", err
	}

	return bz.String(), nil
}

// Encrypt the given privKey with the passphrase using a randomly
// generated salt and the AES-256 GCM cipher
func encryptPrivKey(privKey poktCrypto.PrivateKey, passphrase string) (saltBytes []byte, encBytes []byte, err error) {
	// Get random bytes for salt
	saltBytes = randBytes(16)

	// derive key for encryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	key, err := scrypt.Key([]byte(passphrase), saltBytes, n, r, p, klen)
	if err != nil {
		return nil, nil, err
	}

	//encrypt using AES
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
	var bz bytes.Buffer
	bz.Write([]byte(armourStr))
	dec := gob.NewDecoder(&bz)
	if err = dec.Decode(&armouredKey); err != nil {
		return nil, err
	}

	// check the ArmouredKey for the correct parameters on kdf and salt
	if armouredKey.Kdf != "scrypt" {
		return nil, fmt.Errorf("Unrecognized KDF type: %v", armouredKey.Kdf)
	}
	if armouredKey.Salt == "" {
		return nil, fmt.Errorf("Missing salt bytes")
	}

	//decoding the salt
	saltBytes, err := hex.DecodeString(armouredKey.Salt)
	if err != nil {
		return nil, fmt.Errorf("Error decoding salt: %v", err.Error())
	}

	//decoding the "armoured" ciphertext stored in base64
	encBytes, err := base64.StdEncoding.DecodeString(armouredKey.CipherText)
	if err != nil {
		return nil, fmt.Errorf("Error decoding ciphertext: %v", err.Error())
	}

	//decrypt the actual privkey with the parameters
	privKey, err = decryptPrivKey(saltBytes, encBytes, passphrase)

	return privKey, err
}

// Decrypt the AES-256 GCM encrypted bytes using the passphrase given
func decryptPrivKey(saltBytes []byte, encBytes []byte, passphrase string) (privKey poktCrypto.PrivateKey, err error) {
	// derive key for decryption
	key, err := scrypt.Key([]byte(passphrase), saltBytes, n, r, p, klen)
	if err != nil {
		return nil, err
	}

	//decrypt using AES
	privKeyBytes, err := decryptAESGCM(key, encBytes)
	if err != nil {
		return nil, err
	}

	// Get private key from decrypted bytes
	privKeyBytes, _ = hex.DecodeString(string(privKeyBytes))
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
	if err != nil {
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
