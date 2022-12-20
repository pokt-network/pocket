package converters

import (
	"fmt"
	"math/big"
)

const (
	DefaultDenomination = 10
)

func StringToBigInt(s string) (*big.Int, error) {
	b := big.Int{}
	i, ok := b.SetString(s, DefaultDenomination)
	if !ok {
		return nil, fmt.Errorf("unable to SetString() with base 10")
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(DefaultDenomination)
}
