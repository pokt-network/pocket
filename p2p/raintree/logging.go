package raintree

import (
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/utils"
)

// logStream logs the incoming stream and its scope stats
func (rtr *rainTreeRouter) logStream(stream libp2pNetwork.Stream) {
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
func (rtr *rainTreeRouter) logStreamScopeStats(stream libp2pNetwork.Stream) {
	if err := utils.LogScopeStatFactory(
		&logger.Global.Logger,
		"stream scope (read-side)",
	)(stream.Scope()); err != nil {
		rtr.logger.Debug().Err(err).Msg("logging stream scope stats")
	}
}
