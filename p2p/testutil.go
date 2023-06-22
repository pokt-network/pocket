//go:build test

package p2p

import (
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// WithHost associates an existing (i.e. "started") libp2p `host.Host`
// with this module, instead of creating a new one on `#Start()`.
// Primarily intended for testing.
func WithHost(host libp2pHost.Host) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*p2pModule)
		if ok {
			mod.host = host
			mod.logger.Debug().Msg("using host provided via `WithHost`")
		}
	}
}

func WithUnstakedActorRouter(router typesP2P.Router) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*p2pModule)
		if ok {
			mod.unstakedActorRouter = router
		}
	}
}

func WithStakedActorRouter(router typesP2P.Router) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*p2pModule)
		if ok {
			mod.stakedActorRouter = router
		}
	}
}
