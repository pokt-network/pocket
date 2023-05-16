//go:build test

package p2p

import libp2pNetwork "github.com/libp2p/go-libp2p/core/network"

// P2PModule exports the `p2pModule` type for use in tests
type P2PModule = p2pModule

// HandleStream exports the `handleStream` method for use in tests
func (m *p2pModule) HandleStream(stream libp2pNetwork.Stream) {
	m.handleStream(stream)
}
