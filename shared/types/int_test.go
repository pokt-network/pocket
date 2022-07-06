package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBigIntToString(t *testing.T) {
	bigOriginal := big.NewInt(10)
	bigIntString := BigIntToString(bigOriginal)
	bigIntAfter, err := StringToBigInt(bigIntString)
	require.NoError(t, err)
	if bigIntAfter.Cmp(bigOriginal) != 0 {
		t.Fatal("unequal after conversion")
	}
}
