//go:build test

package p2p

import (
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/pokt-network/pocket/p2p/background"

	"github.com/pokt-network/pocket/p2p/raintree"
)

// P2PModule exports the `p2pModule` type for use in tests
type P2PModule = p2pModule

func (m *p2pModule) GetHost() libp2pHost.Host {
	return m.host
}

// GetRainTreeRouter returns the `RainTreeRouter` for use in integration tests
func (m *p2pModule) GetRainTreeRouter() *raintree.RainTreeRouter {
	return m.stakedActorRouter.(*raintree.RainTreeRouter)
}

// GetBackgroundRouter returns the `BackgroundRouter` for use in integration tests
func (m *p2pModule) GetBackgroundRouter() *background.BackgroundRouter {
	return m.unstakedActorRouter.(*background.BackgroundRouter)
}
