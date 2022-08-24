package types

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
)

const (
	HeightNotUsed = int64(-1)
	EmptyString   = ""
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
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, ErrStringToBigInt()
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(10)
}

func BigIntLessThan(a, b *big.Int) bool {
	if a.Cmp(b) == -1 {
		return true
	}
	return false
}

func Int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}
