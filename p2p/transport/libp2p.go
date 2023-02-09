package transport

import (
	"github.com/libp2p/go-libp2p/core/network"
	"io"

	"github.com/pokt-network/pocket/p2p/types"
)

type libP2PTransport struct {
	stream network.Stream
}

var _ types.Transport = &libP2PTransport{}

func NewLibP2PTransport(stream network.Stream) types.Transport {
	return &libP2PTransport{
		stream: stream,
	}
}

func (transport *libP2PTransport) IsListener() bool {
	// NB: libp2p streams are bi-directional
	return true
}

func (transport *libP2PTransport) Read() ([]byte, error) {
	return io.ReadAll(transport.stream)
}

func (transport *libP2PTransport) Write(data []byte) error {
	_, err := transport.stream.Write(data)
	return err
}

func (transport *libP2PTransport) Close() error {
	return transport.stream.Close()
}
