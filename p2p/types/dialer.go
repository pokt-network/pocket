package types

import (
	"context"
	"net"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
	DialContext(c context.Context, network, address string) (net.Conn, error)
}

func NewDialer() Dialer {
	return &net.Dialer{}
}
