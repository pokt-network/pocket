package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// PoktAccountPathFormat used for HD key derivation where 635 is the SLIP-0044 coin type
	// To be used with fmt.Sprintf to generate child keys
	PoktAccountPathFormat = "m/44'/635'/%d'"
	// firstHardenedKey is the index for the first hardened key for ed25519 keys
	firstHardenedKey = uint32(0x80000000)
	maxHardenedKey   = ^uint32(0)
	MaxChildKeyIndex = maxHardenedKey - firstHardenedKey
	// As defined in: https://github.com/satoshilabs/slips/blob/master/slip-0010.md#master-key-generation
	seedModifier = "ed25519 seed"
)

var (
	ErrInvalidPath  = fmt.Errorf("invalid BIP-44 derivation path")
	ErrNoDerivation = fmt.Errorf("no derivation for an ed25519 key is possible")
	pathRegex       = regexp.MustCompile(`m(/\d+')+$`)
)

type slipKey struct {
	SecretKey []byte
	ChainCode []byte
}

// Derives a key from a BIP-44 path and a seed only operating on hardened ed25519 keys
func DeriveKeyFromPath(path string, seed []byte) (KeyPair, error) {
	segments, err := pathToSegments(path)
	if err != nil {
		return nil, err
	}

	key, err := newMasterKey(seed)
	if err != nil {
		return nil, err
	}

	for _, i32 := range segments {
		// Force hardened keys
		if i32 > MaxChildKeyIndex {
			return nil, fmt.Errorf("hardened key index too large, max: %d, got: %d", MaxChildKeyIndex, i32)
		}
		i := i32 + firstHardenedKey

		// Derive the correct child until final segment
		key, err = key.deriveChild(i)
		if err != nil {
			return nil, err
		}
	}

	return key.convertToKeypair()
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
	// Check i>=2^31 or else no public key can be derived for the ed25519 key
	if i < firstHardenedKey {
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

// Check the BIP-44 path provided is valid and return the []uint32 segments it contains
func pathToSegments(path string) ([]uint32, error) {
	// Check whether the path is valid
	if !pathRegex.MatchString(path) {
		return nil, ErrInvalidPath
	}

	// Split into segments and check for valid uint32 types
	segments := make([]uint32, 0)
	segs := strings.Split(path, "/")
	for _, seg := range segs[1:] {
		ui64, err := strconv.ParseUint(strings.TrimRight(seg, "'"), 10, 32)
		if err != nil {
			return nil, err
		}
		segments = append(segments, uint32(ui64))
	}

	return segments, nil
}
