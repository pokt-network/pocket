package path

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	identifierPrefix  = "#"
	identifierCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._+-#[]<>"
	invalidIdChars    = "/"

	// lengths for identifiers are measured in bytes
	defaultMinIdLength    = 32 // use 32 bytes as a default minimum length to avoid collisions
	defaultMaxIdLength    = 64
	minClientIdLength     = 9
	minConnectionIdLength = 10
	minChannelIdLength    = 8
	minPortIdLength       = 2
	maxPortIdLength       = 128
)

var (
	invalidIdMap map[rune]struct{}
	validIdMap   map[rune]struct{}
)

func init() {
	invalidIdMap = make(map[rune]struct{}, 0)
	for _, c := range invalidIdChars {
		invalidIdMap[c] = struct{}{}
	}
	validIdMap = make(map[rune]struct{}, 0)
	for _, c := range identifierCharset {
		validIdMap[c] = struct{}{}
	}
}

// ValidateClientID validates the client identifier string
func ValidateClientID(id string) error {
	return basicValidation(id, minClientIdLength, defaultMaxIdLength)
}

// ValidateConnectionID validates the connection identifier string
func ValidateConnectionID(id string) error {
	return basicValidation(id, minConnectionIdLength, defaultMaxIdLength)
}

// ValidateChannelID validates the channel identifier string
func ValidateChannelID(id string) error {
	return basicValidation(id, minChannelIdLength, defaultMaxIdLength)
}

// ValidatePortID validates the port identifier string
func ValidatePortID(id string) error {
	return basicValidation(id, minPortIdLength, maxPortIdLength)
}

// GenerateClientIdentifier generates a new client identifier
func GenerateClientIdentifier() string {
	return generateNewIdentifier(minClientIdLength, defaultMaxIdLength)
}

// GenerateConnectionIdentifier generates a new connection identifier
func GenerateConnectionIdentifier() string {
	return generateNewIdentifier(minConnectionIdLength, defaultMaxIdLength)
}

// GenerateChannelIdentifier generates a new channel identifier
func GenerateChannelIdentifier() string {
	return generateNewIdentifier(minChannelIdLength, defaultMaxIdLength)
}

// GeneratePortIdentifier generates a new port identifier
func GeneratePortIdentifier() string {
	return generateNewIdentifier(minPortIdLength, maxPortIdLength)
}

// generateNewIdentifier generates a new identifier in the given range
func generateNewIdentifier(min, max int) string {
	return generateNewIdentifierWithSeed(min, max, time.Now().UnixNano())
}

// generateNewIdentifierWithSeed generates a new identifier in the given range with the identifier prefix
// If the random seed provided is 0 it will use the current unix timestamp as the seed
func generateNewIdentifierWithSeed(min, max int, seed int64) string {
	//nolint:gosec // weak random source okay - cryptographically secure randomness not required
	r := rand.New(rand.NewSource(seed))
	localMin := defaultMinIdLength - min
	size := r.Intn(max-len(identifierPrefix)-localMin) + localMin

	b := make([]byte, size)

	for i := range b {
		b[i] = identifierCharset[r.Intn(len(identifierCharset))]
	}

	return identifierPrefix + string(b)
}

// basicValidation performs basic validation on the given identifier
func basicValidation(id string, minLength, maxLength int) error {
	if strings.TrimSpace(id) == "" {
		return coreTypes.ErrIBCInvalidID(id, "cannot be blank")
	}

	length := len(id)
	if length < minLength || length > maxLength {
		return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("length (%d) must be between %d and %d", length, minLength, maxLength))
	}

	if !strings.HasPrefix(id, identifierPrefix) {
		return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("must start with '%s'", identifierPrefix))
	}

	for _, c := range id {
		if _, ok := invalidIdMap[c]; ok {
			return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("cannot contain '%s'", string(c)))
		}
	}

	for _, c := range id {
		if _, ok := validIdMap[c]; !ok {
			return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("contains invalid character '%c'", c))
		}
	}

	return nil
}
