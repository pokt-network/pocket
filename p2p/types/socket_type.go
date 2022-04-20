package types

type SocketType string

const (
	Outbound            SocketType = "outbound"
	Inbound             SocketType = "inbound"
	UndefinedSocketType SocketType = "unspecified"
)
