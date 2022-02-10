package utility

import (
	"pocket/utility/utility/types"
	"math/big"
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

func BigIntLessThan(a, b *big.Int) bool {
	if a.Cmp(b) == -1 {
		return true
	}
	return false
}

func (u *UtilityContext) CalculateUnstakingHeight(unstakingBlocks int64) (unstakingheight int64, err types.Error) {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return types.ZeroInt, err
	}
	return unstakingBlocks + latestHeight, nil
}
