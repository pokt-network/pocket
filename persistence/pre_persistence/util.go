package pre_persistence

import "math/big"

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

func StringToBigInt(s string) (*big.Int, Error) {
	b := big.NewInt(0)
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, ErrStringToBigInt()
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(10)
}
