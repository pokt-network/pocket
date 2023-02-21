package utils

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
	require.Zero(t, bigIntAfter.Cmp(bigOriginal), "unequal after conversion")
}
