//go:build test

package raintree

import (
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/internal/testutil"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
)

// RainTreeRouter exports `rainTreeRouter` for testing purposes.
type RainTreeRouter = rainTreeRouter

type routerHandlerProxyFactory = testutil.ProxyFactory[typesP2P.MessageHandler]

// HandleStream exports `rainTreeRouter#handleStream` for testing purposes.
func (rtr *rainTreeRouter) HandleStream(stream libp2pNetwork.Stream) {
	rtr.UnicastRouter.HandleStream(stream)
}
func (rtr *rainTreeRouter) HandlerProxy(
	t gocuke.TestingT,
	handlerProxyFactory routerHandlerProxyFactory,
) {
	t.Helper()

	// pass original handler to proxy factory & replace it with the proxy
	origHandler := rtr.handler
	rtr.handler = handlerProxyFactory(origHandler)
}
