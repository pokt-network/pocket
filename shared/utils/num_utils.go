package utils

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

const (
	numericBase = 10
)

func StringToBigInt(s string) (*big.Int, error) {
	b := big.Int{}
	i, ok := b.SetString(s, numericBase)
	if !ok {
		return nil, fmt.Errorf("unable to SetString() with base 10")
	}
	return i, nil
}

func StringToBigFloat(s string) (*big.Float, error) {
	b := big.Float{}
	f, ok := b.SetString(s)
	if !ok {
		return nil, fmt.Errorf("unable to SetString() on float")
	}
	return f, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(numericBase)
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
