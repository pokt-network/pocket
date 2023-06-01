//go:build test

package raintree

import (
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	generics_testutil "github.com/pokt-network/pocket/internal/testutil/generics"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/regen-network/gocuke"
)

// RainTreeRouter exports `rainTreeRouter` for testing purposes.
type RainTreeRouter = rainTreeRouter

// TOOD_THIS_COMMIT: move & dedup
type routerHandlerProxyFactory = generics_testutil.ProxyFactory[typesP2P.RouterHandler]

// HandleStream exports `rainTreeRouter#handleStream` for testing purposes.
func (rtr *rainTreeRouter) HandleStream(stream libp2pNetwork.Stream) {
	rtr.handleStream(stream)
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
