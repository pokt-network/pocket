package unicast

import (
	"io"
	"time"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"

	"github.com/pokt-network/pocket/p2p/config"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

// TECHDEBT(#629): configure timeouts. Consider security exposure vs. real-world conditions.
// TECHDEBT(#629): parameterize and expose via config.
// readStreamTimeout is the duration to wait for a read operation on a
// stream to complete, after which the stream is closed ("timed out").
const readStreamTimeout = time.Second * 10

var _ unicastRouterFactory = &UnicastRouter{}

type unicastRouterFactory = modules.FactoryWithConfig[*UnicastRouter, *config.UnicastRouterConfig]

type UnicastRouter struct {
	base_modules.IntegrableModule

	logger *modules.Logger
	host   libp2pHost.Host
	// messageHandler is the function to call when a message is received.
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	messageHandler typesP2P.MessageHandler
	peerHandler    func(peer typesP2P.Peer) error
}

func Create(bus modules.Bus, cfg *config.UnicastRouterConfig) (*UnicastRouter, error) {
	return new(UnicastRouter).Create(bus, cfg)
}

func (*UnicastRouter) Create(bus modules.Bus, cfg *config.UnicastRouterConfig) (*UnicastRouter, error) {
	if err := cfg.IsValid(); err != nil {
		return nil, err
	}

	rtr := &UnicastRouter{
		logger:         cfg.Logger,
		host:           cfg.Host,
		messageHandler: cfg.MessageHandler,
		peerHandler:    cfg.PeerHandler,
	}
	rtr.SetBus(bus)

	// Don't handle incoming streams in client debug mode.
	if !rtr.isClientDebugMode() {
		rtr.host.SetStreamHandler(cfg.ProtocolID, rtr.handleStream)
	}

	return rtr, nil
}

// handleStream ensures the peerstore contains the remote peer and then reads
// the incoming stream in a new go routine.
func (rtr *UnicastRouter) handleStream(stream libp2pNetwork.Stream) {
	rtr.logger.Debug().Msg("handling incoming stream")
	peer, err := utils.PeerFromLibp2pStream(stream)
	if err != nil {
		rtr.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("parsing remote peer identity")

		if err = stream.Reset(); err != nil {
			rtr.logger.Error().Err(err).Msg("resetting stream")
		}
		return
	}

	if err := rtr.peerHandler(peer); err != nil {
		rtr.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("adding remote peer to router")
	}

	go rtr.readStream(stream)
}

// readStream reads the message bytes out of the incoming stream and passes it to
// the configured `rtr.messageHandler`. Intended to be called in a go routine.
func (rtr *UnicastRouter) readStream(stream libp2pNetwork.Stream) {
	// Time out if no data is sent to free resources.
	// NB: tests using libp2p's `mocknet` rely on this not returning an error.
	if err := stream.SetReadDeadline(newReadStreamDeadline()); err != nil {
		// `SetReadDeadline` not supported by `mocknet` streams.
		rtr.logger.Error().Err(err).Msg("setting stream read deadline")
	}

	// log incoming stream
	rtr.logStream(stream)

	// read stream
	messageBz, err := io.ReadAll(stream)
	if err != nil {
		rtr.logger.Error().Err(err).Msg("reading from stream")
		if err := stream.Reset(); err != nil {
			rtr.logger.Error().Err(err).Msg("resetting stream (read-side)")
		}
		return
	}

	// done reading; reset to signal this to remote peer
	// NB: failing to reset the stream can easily max out the number of available
	// network connections on the receiver's side.
	if err := stream.Reset(); err != nil {
		rtr.logger.Error().Err(err).Msg("resetting stream (read-side)")
	}

	if err := rtr.messageHandler(messageBz); err != nil {
		rtr.logger.Error().Err(err).Msg("handling message")
		return
	}
}

// isClientDebugMode returns the value of `ClientDebugMode` in the base config
func (rtr *UnicastRouter) isClientDebugMode() bool {
	return rtr.GetBus().GetRuntimeMgr().GetConfig().ClientDebugMode
}

// newReadStreamDeadline returns a future deadline
// based on the read stream timeout duration.
func newReadStreamDeadline() time.Time {
	return time.Now().Add(readStreamTimeout)
}
