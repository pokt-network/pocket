//go:build test

package raintree

import libp2pNetwork "github.com/libp2p/go-libp2p/core/network"

// RainTreeRouter exports `rainTreeRouter` for testing purposes.
type RainTreeRouter = rainTreeRouter

// HandleStream exports `rainTreeRouter#handleStream` for testing purposes.
func (rtr *rainTreeRouter) HandleStream(stream libp2pNetwork.Stream) {
	rtr.handleStream(stream)
}
