package types

const (
	MillionInt       = 1000000
	ZeroInt          = 0
	HeightNotUsed    = 0
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
	MaxPort          = 65535
)

type SessionNode struct {
	Address    []byte
	ServiceUrl string
}
