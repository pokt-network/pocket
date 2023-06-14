package crypto

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

const (
	// Encryption params
	kdf                = "scrypt"
	randBz             = 16
	SecretBoxNonceSize = 24
	// Scrypt params
	n    = 32768 // CPU/memory cost param; power of 2 greater than 1
	r    = 8     // r * p < 2³⁰
	p    = 1     // r * p < 2³⁰
	klen = 32    // bytes
)

// Errors
var (
	ErrorWrongPassphrase = errors.New("cannot decrypt private key: wrong passphrase")
)

// Armoured Private Key struct with fields to unarmour it later
type armouredKey struct {
	Kdf        string `json:"kdf"`
	Salt       string `json:"salt"`
	Hint       string `json:"hint"`
	CipherText string `json:"ciphertext"`
}

// Generate new armoured private key struct with parameters for unarmouring
func newArmouredKey(kdf, salt, hint, cipherText string) armouredKey {
	return armouredKey{
		Kdf:        kdf,
		Salt:       salt,
		Hint:       hint,
		CipherText: cipherText,
	}
}

// Encrypt the given privKey with the passphrase, armour it by encoding the encrypted
// []byte into base64, and convert into a json string with the parameters for unarmouring
func encryptArmourPrivKey(privKey PrivateKey, passphrase, hint string) (string, error) {
	// Encrypt privKey using SecretBox cipher
	saltBz, encBz, err := encryptPrivKey(privKey, passphrase)
	if err != nil {
		return "", err
	}

	// Armour encrypted bytes by encoding into Base64
	armourStr := base64.RawStdEncoding.EncodeToString(encBz)

	// Create ArmouredKey object so can unarmour later
	armoured := newArmouredKey(kdf, hex.EncodeToString(saltBz), hint, armourStr)

	// Encode armoured struct into []byte
	js, err := json.Marshal(armoured)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// Encrypt the given privKey with the passphrase using a randomly
// generated salt and the SecretBox cipher
func encryptPrivKey(privKey PrivateKey, passphrase string) (saltBz, encBz []byte, err error) {
	// Get random bytes for salt
	saltBz = randBytes(randBz)

	// Derive key for encryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	encryptionKey, err := scrypt.Key([]byte(passphrase), saltBz, n, r, p, klen)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt using SecretBox
	privKeyHexString := privKey.String()
	encBz, err = encryptCipher(encryptionKey, []byte(privKeyHexString))
	if err != nil {
		return nil, nil, err
	}

	return saltBz, encBz, nil
}

// Unarmor and decrypt the private key using the passphrase provided
func unarmourDecryptPrivKey(armourStr, passphrase string) (privKey PrivateKey, err error) {
	// Decode armourStr back into ArmouredKey struct
	ak := armouredKey{}
	err = json.Unmarshal([]byte(armourStr), &ak)
	if err != nil {
		return nil, err
	}

	// Check the ArmouredKey for the correct parameters on kdf and salt
	if ak.Kdf != kdf {
		return nil, fmt.Errorf("unrecognized KDF type: %v", ak.Kdf)
	}
	if ak.Salt == "" {
		return nil, fmt.Errorf("missing salt bytes")
	}

	// Decoding the salt
	saltBz, err := hex.DecodeString(ak.Salt)
	if err != nil {
		return nil, fmt.Errorf("error decoding salt: %v", err.Error())
	}

	// Decoding the "armoured" ciphertext stored in base64
	encBz, err := base64.RawStdEncoding.DecodeString(ak.CipherText)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %v", err.Error())
	}

	// Decrypt the actual privkey with the parameters
	privKey, err = decryptPrivKey(saltBz, encBz, passphrase)

	return privKey, err
}

// Decrypt the SecretBox encrypted bytes using the passphrase given
func decryptPrivKey(saltBz, encBz []byte, passphrase string) (PrivateKey, error) {
	// Derive key for decryption, see: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
	encryptionKey, err := scrypt.Key([]byte(passphrase), saltBz, n, r, p, klen)
	if err != nil {
		return nil, err
	}

	// Decrypt using SecretBox
	privKeyRawHexBz, err := decryptCipher(encryptionKey, encBz)
	if err != nil {
		return nil, err
	}
	bz, err := hex.DecodeString(string(privKeyRawHexBz))
	if err != nil {
		return nil, err
	}

	// Get private key from decrypted bytes
	pk, err := NewPrivateKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// Encrypt using NaCl SecretBox (XSalsa20 + Poly1305)
func encryptCipher(key, plaintext []byte) ([]byte, error) {
	var nonce [SecretBoxNonceSize]byte
	if _, err := io.ReadFull(crand.Reader, nonce[:]); err != nil {
		return nil, err
	}
	var secretKey [klen]byte
	copy(secretKey[:], key)
	encBz := secretbox.Seal(nonce[:], plaintext, &nonce, &secretKey)
	return encBz, nil
}

// Decrypt using NaCl SecretBox
func decryptCipher(key, encBz []byte) ([]byte, error) {
	var secretKey [klen]byte
	copy(secretKey[:], key)
	var decryptNonce [SecretBoxNonceSize]byte
	copy(decryptNonce[:], encBz[:SecretBoxNonceSize])
	decrypted, ok := secretbox.Open(nil, encBz[SecretBoxNonceSize:], &decryptNonce, &secretKey)
	if !ok {
		return nil, ErrorWrongPassphrase
	}

	return decrypted, nil
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
