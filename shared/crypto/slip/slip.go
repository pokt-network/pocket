package slip

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pokt-network/pocket/shared/crypto"
)

const (
	// PoktAccountPathFormat used for HD key derivation where 635 is the SLIP-0044 coin type
	// To be used with fmt.Sprintf to generate child keys
	// Ref: https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	PoktAccountPathFormat = "m/44'/635'/%d'" // m/purpose'/coin_type'/account_idx'
	// As defined in: https://github.com/satoshilabs/slips/blob/master/slip-0010.md#master-key-generation
	seedModifier = "ed25519 seed"
	// Hardened key values, as defined in: https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#extended-keys
	firstHardenedKeyIndex = uint32(1 << 31) // 2^31; [0, 2^31) are non-hardened keys which are not supported for ed25519
	maxHardenedKeyIndex   = ^uint32(0)      // 2^32-1
	maxChildKeyIndex      = maxHardenedKeyIndex - firstHardenedKeyIndex
)

var (
	ErrNoDerivation = fmt.Errorf("no derivation for a hardened ed25519 key is possible")
	ErrInvalidPath  = fmt.Errorf("invalid BIP-44 derivation path")
	pathRegex       = regexp.MustCompile(`m(/\d+')+$`)
)

type slipKey struct {
	secretKey []byte
	chainCode []byte
}

// Derives a master key from the seed provided and returns the child at the correct index
func DeriveChild(path string, seed []byte) (crypto.KeyPair, error) {
	// Break down path into uint32 segments
	segments, err := pathToSegments(path)
	if err != nil {
		return nil, err
	}

	// Initialise a master key from seed to start child generation
	key, err := newMasterKey(seed)
	if err != nil {
		return nil, err
	}

	// Iterate over segments in path regenerating the child key until correct
	// Enforce that both the purpose and coin type paths are using hardened keys
	for _, i32 := range segments {
		// Force hardened keys
		if i32 > maxChildKeyIndex {
			return nil, fmt.Errorf("hardened key index too large, max: %d, got: %d", maxChildKeyIndex, i32)
		}
		keyIdx := i32 + firstHardenedKeyIndex

		// Derive the correct child until final segment
		key, err = key.deriveChild(keyIdx)
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
		secretKey: sum[:32],
		chainCode: sum[32:],
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
	data = append(data, k.secretKey...)
	data = append(data, iBytes...)

	hmacHash := hmac.New(sha512.New, k.chainCode)
	if _, err := hmacHash.Write(data); err != nil {
		return nil, err
	}

	// Convert hash to []byte
	sum := hmacHash.Sum(nil)

	// Create SLIP-0010 Child key
	child := &slipKey{
		secretKey: sum[:32],
		chainCode: sum[32:],
	}
	return child, nil
}

func (k *slipKey) convertToKeypair() (crypto.KeyPair, error) {
	return crypto.CreateNewKeyFromSeed(k.secretKey, "", "")
}

// Check the BIP-44 path provided is valid and return the []uint32 segments it contains
func pathToSegments(path string) ([]uint32, error) {
	// Master path exception
	if path == "m" {
		return []uint32{}, nil
	}

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
