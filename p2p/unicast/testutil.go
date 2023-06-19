//go:build test

package unicast

import libp2pNetwork "github.com/libp2p/go-libp2p/core/network"

// HandleStream exports `unicastRouter#handleStream` for testing purposes.
func (rtr *UnicastRouter) HandleStream(stream libp2pNetwork.Stream) {
	rtr.handleStream(stream)
}
