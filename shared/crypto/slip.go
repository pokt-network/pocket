package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
)

const (
	// As defined in: https://github.com/satoshilabs/slips/blob/master/slip-0010.md#master-key-generation
	seedModifier = "ed25519 seed"
	// Hardened key values, as defined in: https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#extended-keys
	firstHardenedKeyIndex = uint32(2147483648) // 2^31
	maxHardenedKeyIndex   = ^uint32(0)         // 2^32-1
	maxChildKeyIndex      = maxHardenedKeyIndex - firstHardenedKeyIndex
)

var (
	ErrNoDerivation = fmt.Errorf("no derivation for an ed25519 key is possible")
)

type slipKey struct {
	SecretKey []byte
	ChainCode []byte
}

// Derives a master key from the seed provided and returns the child at the correct index
func DeriveChild(index uint32, seed []byte) (KeyPair, error) {
	masterKey, err := newMasterKey(seed)
	if err != nil {
		return nil, err
	}

	// Allow index usage from 0 while using hardened keys
	if index > maxChildKeyIndex {
		return nil, fmt.Errorf("child index is greater than max hardened ed25519 key index: got %d, max %d", index, maxChildKeyIndex)
	}

	// Force hardened keys
	index += firstHardenedKeyIndex

	childKey, err := masterKey.deriveChild(index)
	if err != nil {
		return nil, err
	}

	return childKey.convertToKeypair()
}

// Reference: https://github.com/satoshilabs/slips/blob/master/slip-0010.md#master-key-generation
func newMasterKey(seed []byte) (*slipKey, error) {
	// Create HMAC hash from curve and seed
	hmacHash := hmac.New(sha512.New, []byte(seedModifier))
	if _, err := hmacHash.Write(seed); err != nil {
		return nil, err
	}

	// Convert hash to []byte
	sum := hmacHash.Sum(nil)

	// Create SLIP-0010 Master key
	key := &slipKey{
		SecretKey: sum[:32],
		ChainCode: sum[32:],
	}

	return key, nil
}

// Reference: https://github.com/satoshilabs/slips/blob/master/slip-0010.md#private-parent-key--private-child-key
func (k *slipKey) deriveChild(i uint32) (*slipKey, error) {
	if i < firstHardenedKeyIndex {
		return nil, ErrNoDerivation
	}

	// i is a hardened child compute the HMAC hash of the key
	data := []byte{0x0}
	iBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(iBytes, i)
	data = append(data, k.SecretKey...)
	data = append(data, iBytes...)

	hmacHash := hmac.New(sha512.New, k.ChainCode)
	if _, err := hmacHash.Write(data); err != nil {
		return nil, err
	}

	// Convert hash to []byte
	sum := hmacHash.Sum(nil)

	// Create SLIP-0010 Child key
	child := &slipKey{
		SecretKey: sum[:32],
		ChainCode: sum[32:],
	}
	return child, nil
}

func (k *slipKey) convertToKeypair() (KeyPair, error) {
	// Generate PrivateKey interface form secret key
	reader := bytes.NewReader(k.SecretKey)
	privKey, err := GeneratePrivateKeyWithReader(reader)
	if err != nil {
		return nil, err
	}

	// Armour and encrypt private key into JSON string
	armouredStr, err := encryptArmourPrivKey(privKey, "", "") // No passphrase or hint as they depend on the master key
	if err != nil {
		return nil, err
	}

	// Return KeyPair interface
	return &encKeyPair{
		PublicKey:     privKey.PublicKey(),
		PrivKeyArmour: armouredStr,
	}, nil
}
