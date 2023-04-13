package utils

import (
	"net"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/rs/zerolog"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
)

type scopeCallback func(scope network.ResourceScope) error

// LogScopeStatFactory returns a function which prints the given scope stat fields
// to the debug level of the provided logger.
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.27.0/core/network#ScopeStat)
// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.27.0/core/network#ResourceScope)
// TECHDEBT: would prefer receive a pocket logger object instead.
// Typical calls would pass either `logger.Global` or a `*modules.Logger` which
// are disparate types.
func LogScopeStatFactory(logger *zerolog.Logger, msg string) scopeCallback {
	return func(scope network.ResourceScope) error {
		stat := scope.Stat()
		logger.Debug().Fields(map[string]any{
			"InboundConns":    stat.NumConnsInbound,
			"OutboundConns":   stat.NumConnsOutbound,
			"InboundStreams":  stat.NumStreamsInbound,
			"OutboundStreams": stat.NumStreamsOutbound,
		}).Msg(msg)
		return nil
	}
}

func LogOutgoingMsg(logger *modules.Logger, hostname string, peer types.Peer) {
	msg := "OUTGOING MSG"
	logMessage(logger, msg, hostname, peer)
}

func LogIncomingMsg(logger *modules.Logger, hostname string, peer types.Peer) {
	msg := "INCOMING MSG"
	logMessage(logger, msg, hostname, peer)
}

func logMessage(logger *modules.Logger, msg, hostname string, peer types.Peer) {
	remoteHostname, _, err := net.SplitHostPort(peer.GetServiceURL())
	if err != nil {
		logger.Debug().Err(err).
			Str("serviceURL", peer.GetServiceURL()).
			Msg("parsing remote service URL")
		return
	}

	logger.Debug().Fields(map[string]any{
		"local_hostname":  hostname,
		"remote_hostname": remoteHostname,
	}).Msg(msg)
}
