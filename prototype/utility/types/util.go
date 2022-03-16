package types

import (
	"crypto/rand"
	"math/big"
)

const (
	ZeroInt          = 0
	EmptyString      = ""
	HttpsPrefix      = "https://"
	HttpPrefix       = "http://"
	Colon            = ":"
	Period           = "."
	InvalidURLPrefix = "the url must start with http:// or https://"
	PortRequired     = "a port is required"
	NonNumberPort    = "invalid port, cant convert to integer"
	PortOutOfRange   = "invalid port, out of valid port range"
	NoPeriod         = "must contain one '.'"
)

type Pagination struct {
}

type Page struct {
	Size int
	Skip int
	Sort string
}

func RandBigInt() *big.Int {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(max, big.NewInt(1))
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
