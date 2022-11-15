package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// Utility functions

// Print out key in indented JSON format
func printKey(keystore key) error {
	output, err := json.MarshalIndent(keystore, "", "\t")
	if err != nil {
		return err
	}
	log.Printf("%s\n", output)

	return nil
}

// Encryption function that takes a 32 bytes hex string key
func encrypt(stringToEncrypt string, keyString string) (string, error) {
	var err error

	// convert key and plaintext to bytes
	var key, plaintext []byte
	if key, err = hex.DecodeString(keyString); err != nil {
		return "", err
	}
	plaintext = []byte(stringToEncrypt)

	if len(key) != 32 {
		return "", errors.New("key size much be 32 bytes for AES-256 security level")
	}

	// create a new cipher Block from the key
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return "", err
	}

	// create a new GCM (Galois Counter Mode)
	// https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	var aesGCM cipher.AEAD
	if aesGCM, err = cipher.NewGCM(block); err != nil {
		return "", err
	}

	// create nonce from GCM (nonce size = 12)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// encrypt the data using the AEA GCM Seal function
	enc := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return hex.EncodeToString(enc), nil
}

// Decryption function that takes a 32 bytes hex string key
func decrypt(encryptedString string, keyString string) (string, error) {
	var err error

	// convert key and encrypted data to bytes
	var key, enc []byte
	if key, err = hex.DecodeString(keyString); err != nil {
		return "", err
	}
	if enc, err = hex.DecodeString(encryptedString); err != nil {
		return "", err
	}

	if len(key) != 32 {
		return "", errors.New("key size much be 32 bytes for AES-256 security level")
	}

	// create a new cipher Block from the key
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return "", err
	}

	// create a new GCM (Galois Counter Mode)
	var aesGCM cipher.AEAD
	if aesGCM, err = cipher.NewGCM(block); err != nil {
		return "", err
	}

	// get nonce size and extract nonce from the prefix of the ciphertext
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	// decrypt the data
	var plaintext []byte
	if plaintext, err = aesGCM.Open(nil, nonce, ciphertext, nil); err != nil {
		return "", err
	}

	return string(plaintext), nil
}

/* Generating strong random passphrase for user
In most use cases `generateKeyFromPassPhrase` should be sufficient

Warning: this passphrase is not saved by default.
Warning: recommend to store this pass phase in safe vault for users in case they lost it.

*/
func generateRandomKey() (string, error) {
	// generate a random 32 byte key for AES-256
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Keep the key as a secrete! (User only need to remember their passphrase, could be "")
	passphrase := hex.EncodeToString(bytes)
	log.Printf("Generated user passphrase: %s\n", passphrase)
	log.Println("Warning: write this passphrase down!")
	log.Println("Warning: this passphrase will disappear!")
	return passphrase, nil
}

/* Generate a 32 byte hash digest as key based on the user provided pass phrase.
   SHA3-256 only has 128-bit collision resistance, because its output length is 32 bytes.

- Input
	- stringToHash: string of any length

*/
func generateKeyFromPassPhrase(passphrase string) (string, error) {
	// generate a strong random secret key that is at least 32 bytes long
	buf := []byte(passphrase)
	h := cryptoPocket.SHA3Hash(buf)

	// check key length
	if len(h) != 32 {
		return "", errors.New("key generation error: hash digest must be 32 bytes long")
	}

	// return the 32 bytes key hex string
	return hex.EncodeToString(h), nil
}

// Output logs for essential mnemonic and key information
func logInfo(keystore key, mnemonic string) error {
	log.Printf("Make sure you have the mnemonic saved in a safe place!")
	log.Printf("\t%v\n", mnemonic)
	log.Println("Key information (encrypted private key string)")

	var keyJSON []byte
	var err error
	if keyJSON, err = json.MarshalIndent(keystore, "", "\t"); err != nil {
		return err
	}
	log.Printf("%s\n", keyJSON)

	return nil
}
