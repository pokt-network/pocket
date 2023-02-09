package converters

import (
	"encoding/binary"
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
		return nil, fmt.Errorf("unable to SetString() with base 10")
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(defaultDenomination)
}

func BigIntLessThan(a, b *big.Int) bool {
	return a.Cmp(b) == -1
}

func HeightFromBytes(heightBz []byte) uint64 {
	return binary.LittleEndian.Uint64(heightBz)
}

func HeightToBytes(height uint64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, height)
	return heightBytes
}
