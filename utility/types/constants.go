package types

// CLEANUP: Consider moving these into a shared location or eliminating altogether
const (
	MillionInt       = 1000000
	ZeroInt          = 0
	HeightNotUsed    = int64(-1)
	EmptyString      = ""
	httpsPrefix      = "https://"
	httpPrefix       = "http://"
	colon            = ":"
	period           = "."
	invalidURLPrefix = "the url must start with http:// or https://"
	portRequired     = "a port is required"
	NonNumberPort    = "invalid port, cant convert to integer"
	PortOutOfRange   = "invalid port, out of valid port range"
	NoPeriod         = "must contain one '.'"
	maxPort          = 65535
)
