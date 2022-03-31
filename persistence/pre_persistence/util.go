package pre_persistence

import (
	"encoding/binary"
	"math/big"

	"github.com/pokt-network/pocket/shared/types"
)

const (
	ZeroInt     = 0
	EmptyString = ""
)

func StringToBigInt(s string) (*big.Int, types.Error) {
	b := &big.Int{}
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, types.ErrStringToBigInt()
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(10)
}

func Int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}
