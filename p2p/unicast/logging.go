package unicast

import (
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/utils"
)

// TECHDEBT(#830): it would be nice to have at least one more degree of freedom with which
// to limit logging in areas where it is known to be excessive / high frequency.
// Especially applicable to debug log lines which only contribute in edge cases,
// unusual circumstances, or regressions (e.g. hitting OS resource limits because
// of too many concurrent streams).
//
// This could ultimately be actuated from the CLI via flags, configs, and/or env
// vars. Initially, we could consider coupling to a `--verbose` persistent flag.
//

// logStream logs the incoming stream and its scope stats
func (rtr *UnicastRouter) logStream(stream libp2pNetwork.Stream) {
	rtr.logStreamScopeStats(stream)

	remotePeer, err := utils.PeerFromLibp2pStream(stream)
	if err != nil {
		rtr.logger.Debug().Err(err).Msg("getting remote remotePeer")
	} else {
		utils.LogIncomingMsg(rtr.logger, rtr.getHostname(), remotePeer)
	}
}

// logStreamScopeStats logs the incoming stream's scope stats
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.27.0/core/network#StreamScope)
func (rtr *UnicastRouter) logStreamScopeStats(stream libp2pNetwork.Stream) {
	if err := utils.LogScopeStatFactory(
		&logger.Global.Logger,
		"stream scope (read-side)",
	)(stream.Scope()); err != nil {
		rtr.logger.Debug().Err(err).Msg("logging stream scope stats")
	}
}
func (rtr *UnicastRouter) getHostname() string {
	return rtr.GetBus().GetRuntimeMgr().GetConfig().P2P.Hostname
}
