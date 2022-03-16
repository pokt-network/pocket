package pre_persistence

import "math/big"

const (
	ZeroInt     = 0
	EmptyString = ""
)

type Pagination struct {
}

type Page struct {
	Size int
	Skip int
	Sort string
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
