package pre_persistence

import (
	"github.com/pokt-network/pocket/shared/types"
	"math/big"
)

const (
	ZeroInt       = 0
	heightNotUsed = 0 // TODO (Andrew) update design, could use -1
	EmptyString   = ""
)

func StringToBigInt(s string) (*big.Int, types.Error) {
	b := big.NewInt(0)
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, types.ErrStringToBigInt()
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(10)
}
