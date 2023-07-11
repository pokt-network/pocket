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

// WithUnstakedActorRouter assigns the given router to the P2P modules
// `#unstakedActor` field, used to communicate between unstaked actors
// and the rest of the network, plus as a redundancy to the staked actor
// router when broadcasting.
func WithUnstakedActorRouter(router typesP2P.Router) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*p2pModule)
		if ok {
			mod.unstakedActorRouter = router
			mod.logger.Debug().Msg("using unstaked actor router provided via `WithUnstakeActorRouter`")
		}
	}
}

// WithStakedActorRouter assigns the given router to the P2P modules'
// `#stakedActor` field, exclusively used to communicate between staked actors.
func WithStakedActorRouter(router typesP2P.Router) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*p2pModule)
		if ok {
			mod.stakedActorRouter = router
			mod.logger.Debug().Msg("using staked actor router provided via `WithStakeActorRouter`")
		}
	}
}
