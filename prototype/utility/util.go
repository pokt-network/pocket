package utility

import (
	"math/big"
	types2 "pocket/utility/types"
)

func StringToBigInt(s string) (*big.Int, types2.Error) {
	b := big.Int{}
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, types2.ErrStringToBigInt()
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

func (u *UtilityContext) CalculateUnstakingHeight(unstakingBlocks int64) (unstakingheight int64, err types2.Error) {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return types2.ZeroInt, err
	}
	return unstakingBlocks + latestHeight, nil
}
