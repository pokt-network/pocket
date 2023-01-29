package types

// CLEANUP: Consider moving these into a shared location or eliminating altogether
const (
	MillionInt       = 1000000
	ZeroInt          = 0
	HeightNotUsed    = int64(-1)
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
