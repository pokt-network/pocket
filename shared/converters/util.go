package converters

import (
	"encoding/binary"
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

func HeightFromBytes(heightBz []byte) uint64 {
	return binary.LittleEndian.Uint64(heightBz)
}

func HeightToBytes(height uint64) []byte {

	heightBytes := make([]byte, 8)
	//binary.LittleEndian.AppendUint64(heightBytes, height)
	binary.LittleEndian.PutUint64(heightBytes, height)
	fmt.Printf("INSIDE BLOCK HEIGHT CONVERTER FOR: %d, byte array size: %d \n", height, len(heightBytes))
	return heightBytes
}
