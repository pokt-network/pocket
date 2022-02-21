package p2p

var (
	MaxInbound           uint = 128
	MaxOutbound          uint = 128
	WireByteHeaderLength int  = 9
	ReadBufferSize       int  = (1024 * 4)
	WriteBufferSize      int  = (1024 * 4)
	ReadDeadlineMs       int  = 400
)

var (
	Protocol = "tcp"
	Address  = "localhost:30303"
)
