package host

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	defaultIdentifierLength = 32
	identifierPrefix        = "#"
	identifierCharset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._+-#[]<>"
	defaultMaxIdLength      = 64
	portMaxIdLength         = 128
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

	if strings.Contains(id, "/") {
		return coreTypes.ErrIBCInvalidID(id, "cannot contain '/'")
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
	return basicValidation(id, 9, defaultMaxIdLength)
}

// ValidateConnectionID validates the connection identifier string
func ValidateConnectionID(id string) error {
	return basicValidation(id, 10, defaultMaxIdLength)
}

// ValidateChannelID validates the channel identifier string
func ValidateChannelID(id string) error {
	return basicValidation(id, 8, defaultMaxIdLength)
}

// ValidatePortID validates the port identifier string
func ValidatePortID(id string) error {
	return basicValidation(id, 2, portMaxIdLength)
}

// generateNewIdentifier generates a new identifier in the given range with the identifier prefix
func generateNewIdentifier(min, max int) string { //nolint:unparam // min is used although always the same
	//nolint:gosec // weak random source okay - cryptographically secure randomness not required
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	size := r.Intn(max-1-min) + min // -1 for the prefix

	b := make([]byte, size)

	for i := range b {
		b[i] = identifierCharset[r.Intn(len(identifierCharset))]
	}

	return identifierPrefix + string(b)
}

// GenerateClientIdentifier generates a new client identifier
func GenerateClientIdentifier() string {
	return generateNewIdentifier(defaultIdentifierLength, defaultMaxIdLength)
}

// GenerateConnectionIdentifier generates a new connection identifier
func GenerateConnectionIdentifier() string {
	return generateNewIdentifier(defaultIdentifierLength, defaultMaxIdLength)
}

// GenerateChannelIdentifier generates a new channel identifier
func GenerateChannelIdentifier() string {
	return generateNewIdentifier(defaultIdentifierLength, defaultMaxIdLength)
}

// GeneratePortIdentifier generates a new port identifier
func GeneratePortIdentifier() string {
	return generateNewIdentifier(defaultIdentifierLength, portMaxIdLength)
}
