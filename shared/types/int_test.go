package types

import (
	"math/big"
	"testing"
)

func TestBigIntToString(t *testing.T) {
	bigOriginal := big.NewInt(10)
	bigIntString := BigIntToString(bigOriginal)
	bigIntAfter, err := StringToBigInt(bigIntString)
	if err != nil {
		t.Fatal(err)
	}
	if bigIntAfter.Cmp(bigOriginal) != 0 {
		t.Fatal("unequal after conversion")
	}
}
