package types

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPools_Address(t *testing.T) {
	tests := []struct {
		name string
		pn   Pools
	}{
		// initializing tests with the custom/edge cases
		{"unspecified", Pools_POOLS_UNSPECIFIED},
		{"invalid", Pools(100)},
	}

	// adding all the real world cases programmatically in order to catch any changes to the enum
	// that must be reflected into the hardcoded values
	for _, pool := range Pools_value {
		tests = append(tests, struct {
			name string
			pn   Pools
		}{
			name: Pools(pool).FriendlyName(),
			pn:   Pools(pool),
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.pn.FriendlyName()
			want, err := convertStringToHex(name)
			require.NoError(t, err)

			if got := tt.pn.Address(); got != want {
				t.Errorf("Pools.Address() = %v, want %v", got, want)
			}
		})
	}
}

// convertStringToHex is the function used to opinionatedly convert a pool name into a valid address
// this is done by encoding to hex and padding to 40 characters with zeros
func convertStringToHex(s string) (string, error) {
	if len(s) == 0 {
		return "", nil
	}
	src := []byte(s)
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	encodedStr := string(dst)

	if len(encodedStr) > 40 {
		return "", errors.New("the resulting string is longer than 40 characters")
	}

	// Add zeros to the end of the encoded string until it reaches 40 characters
	paddedStr := encodedStr + strings.Repeat("0", 40-len(encodedStr))

	return paddedStr, nil
}