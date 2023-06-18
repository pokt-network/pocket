package host

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
	minClientIdLength     = 9
	minConnectionIdLength = 10
	minChannelIdLength    = 8
	minPortIdLength       = 2
	defaultMinIdLength    = 32 // use 32 bytes as a default minimum length to avoid collisions
	defaultMaxIdLength    = 64
	portMaxIdLength       = 128
)

// basicValidation performs basic validation on the given identifier
func basicValidation(id string, minLength, maxLength int) error {
	if strings.TrimSpace(id) == "" {
		return coreTypes.ErrIBCInvalidID(id, "cannot be blank")
	}

	if len(id) < minLength || len(id) > maxLength {
		return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("length must be between %d and %d", minLength, maxLength))
	}

	if !strings.HasPrefix(id, identifierPrefix) {
		return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("must start with '%s'", identifierPrefix))
	}

	for _, c := range invalidIdChars {
		if strings.Contains(id, string(c)) {
			return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("cannot contain '%s'", string(c)))
		}
	}

	for _, c := range id {
		if ok := strings.Contains(identifierCharset, string(c)); !ok {
			return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("contains invalid character '%c'", c))
		}
	}

	return nil
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
	return basicValidation(id, minPortIdLength, portMaxIdLength)
}

// generateNewIdentifier generates a new identifier in the given range with the identifier prefix
// If the random seed provided is 0 it will use the current unix timestamp as the seed
func generateNewIdentifier(min, max int, seed int64) string {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
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

// GenerateClientIdentifier generates a new client identifier
func GenerateClientIdentifier(seed int64) string {
	return generateNewIdentifier(minClientIdLength, defaultMaxIdLength, seed)
}

// GenerateConnectionIdentifier generates a new connection identifier
func GenerateConnectionIdentifier(seed int64) string {
	return generateNewIdentifier(minConnectionIdLength, defaultMaxIdLength, seed)
}

// GenerateChannelIdentifier generates a new channel identifier
func GenerateChannelIdentifier(seed int64) string {
	return generateNewIdentifier(minChannelIdLength, defaultMaxIdLength, seed)
}

// GeneratePortIdentifier generates a new port identifier
func GeneratePortIdentifier(seed int64) string {
	return generateNewIdentifier(minPortIdLength, portMaxIdLength, seed)
}
