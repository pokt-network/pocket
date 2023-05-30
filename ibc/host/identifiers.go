package host

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	identifierPrefix   = "#"
	identifierCharset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._+-#[]<>"
	portMinIdLength    = 2
	channelMinIdLength = 8
	clientMinIdLength  = 9
	connectionMinIdLen = 10
	defaultMaxIdLength = 64
	portMaxIdLength    = 128
)

func basicValidation(id string, minLength, maxLength int) error {
	if strings.TrimSpace(id) == "" {
		return coreTypes.ErrIBCInvalidID(id, "cannot be blank")
	}

	if strings.Contains(id, "/") {
		return coreTypes.ErrIBCInvalidID(id, "cannot contain '/'")
	}

	if len(id) < minLength || len(id) > maxLength {
		return coreTypes.ErrIBCInvalidID(id, fmt.Sprintf("length must be between %d and %d", minLength, maxLength))
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
	return basicValidation(id, clientMinIdLength, defaultMaxIdLength)
}

// ValidateConnectionID validates the connection identifier string
func ValidateConnectionID(id string) error {
	return basicValidation(id, connectionMinIdLen, defaultMaxIdLength)
}

// ValidateChannelID validates the channel identifier string
func ValidateChannelID(id string) error {
	return basicValidation(id, channelMinIdLength, defaultMaxIdLength)
}

// ValidatePortID validates the port identifier string
func ValidatePortID(id string) error {
	return basicValidation(id, portMinIdLength, portMaxIdLength)
}

// generateNewIdentifier generates a new identifier in the given range with the identifier prefix
func generateNewIdentifier(min, max int) string {
	//nolint:gosec // weak random source okay - cryptographically secure randomness not required
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	size := r.Intn(max-1-min) + min // -1 for

	b := make([]byte, size)

	for i := range b {
		b[i] = identifierCharset[r.Intn(len(identifierCharset))]
	}

	return identifierPrefix + string(b)
}

// GenerateClientIdentifier generates a new client identifier
func GenerateClientIdentifier() string {
	return generateNewIdentifier(clientMinIdLength, defaultMaxIdLength)
}

// GenerateConnectionIdentifier generates a new connection identifier
func GenerateConnectionIdentifier() string {
	return generateNewIdentifier(connectionMinIdLen, defaultMaxIdLength)
}

// GenerateChannelIdentifier generates a new channel identifier
func GenerateChannelIdentifier() string {
	return generateNewIdentifier(channelMinIdLength, defaultMaxIdLength)
}

// GeneratePortIdentifier generates a new port identifier
func GeneratePortIdentifier() string {
	return generateNewIdentifier(portMinIdLength, portMaxIdLength)
}
