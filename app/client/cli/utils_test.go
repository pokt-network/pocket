package cli

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func Test_parseEd25519PrivateKeyFromReader_NilInput(t *testing.T) {
	_, err := parseEd25519PrivateKeyFromReader(nil)
	require.Error(t, err)
}

func Test_parseEd25519PrivateKeyFromReader_EmptyByteArray(t *testing.T) {
	_, err := parseEd25519PrivateKeyFromReader(bytes.NewReader([]byte{}))
	require.Error(t, err)
}

func Test_parseEd25519PrivateKeyFromReader_ValidPrivateKey(t *testing.T) {
	validPKString := `"e7760141c2672178b28360a8cf80ff3a9d5fd579990317b9afcb2091426ffe75dc12b26584c057be33fcc8e891a483250581e38fe2bc9d62c1a1341c5e85b667"`
	pk, err := strconv.Unquote(validPKString)
	require.NoError(t, err)

	validPk, err := crypto.NewPrivateKey(pk)
	require.NoError(t, err)

	gotPk, err := parseEd25519PrivateKeyFromReader(strings.NewReader(validPKString))
	require.NoError(t, err)
	if !reflect.DeepEqual(gotPk, validPk) {
		t.Errorf("parseEd25519PrivateKeyFromFile() = %v, want %v", gotPk, validPk)
	}
}
