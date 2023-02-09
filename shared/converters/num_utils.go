package converters

import (
	"fmt"
	"math/big"
)

const (
	defaultDenomination = 10
)

func StringToBigInt(s string) (*big.Int, error) {
	b := big.Int{}
	i, ok := b.SetString(s, defaultDenomination)
	if !ok {
		return nil, fmt.Errorf("unable to SetString(%s) with base %d", s, defaultDenomination)
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(defaultDenomination)
}

func BigIntLessThan(a, b *big.Int) bool {
	return a.Cmp(b) == -1
}
