package types

import (
	"crypto/rand"
	"math/big"
)

const (
	DefaultDenomination = 10
)

var max *big.Int

func init() {
	max = new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(max, big.NewInt(1))
}

func RandBigInt() *big.Int {
	n, _ := rand.Int(rand.Reader, max)
	return n
}

func StringToBigInt(s string) (*big.Int, Error) {
	b := big.Int{}
	i, ok := b.SetString(s, DefaultDenomination)
	if !ok {
		return nil, ErrStringToBigInt()
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(DefaultDenomination)
}

func BigIntLessThan(a, b *big.Int) bool {
	return a.Cmp(b) == -1
}
