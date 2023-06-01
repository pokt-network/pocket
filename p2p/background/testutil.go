//go:build test

package background

import (
	"github.com/pokt-network/pocket/internal/testutil/generics"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/regen-network/gocuke"
)

// BackgroundRouter exports `backgroundRouter` for testing purposes.
type BackgroundRouter = backgroundRouter

// TOOD_THIS_COMMIT: move & dedup
type routerHandlerProxyFactory = generics_testutil.ProxyFactory[typesP2P.RouterHandler]

func (rtr *backgroundRouter) HandlerProxy(
	t gocuke.TestingT,
	handlerProxyFactory routerHandlerProxyFactory,
) {
	t.Helper()

	// pass original handler to proxy factory & replace it with the proxy
	rtr.handler = handlerProxyFactory(rtr.handler)
}

