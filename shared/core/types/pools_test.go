package types

import (
	"bytes"
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
			want, err := convertFriendlyNameToHexBytes(name)
			require.NoError(t, err)

			if got := tt.pn.Address(); !bytes.Equal(got, want) {
				t.Errorf("Pools.Address() = %v, want %v", string(got), string(want))
			}
		})
	}
}

// convertFriendlyNameToHexBytes is the function used to opinionatedly convert a pool name into a valid address
// this is done by encoding to hex and padding to 40 characters with zeros
// TODO: consider doing the same as V0 (https://github.com/pokt-network/pocket-core/blob/a109dfc03a13eec06413bf1eb7d17fe093f96842/x/auth/types/account.go#L320)
func convertFriendlyNameToHexBytes(s string) ([]byte, error) {
	if s == "" {
		return []byte(""), nil
	}
	src := []byte(s)
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	encodedStr := string(dst)

	if len(encodedStr) > 40 {
		return []byte(""), errors.New("the resulting string is longer than 40 characters")
	}

	// Add zeros to the end of the encoded string until it reaches 40 characters
	paddedStr := encodedStr + strings.Repeat("0", 40-len(encodedStr))

	return []byte(paddedStr), nil
}
