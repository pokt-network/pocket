package types

import (
	"fmt"
	"math/big"
)

func StringToBigInt(s string) (*big.Int, error) {
	b := big.Int{}
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("unable to SetString() with base 10")
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(10)
}
