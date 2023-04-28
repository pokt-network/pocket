// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package background

import (
	"context"
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/protocol"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ typesP2P.Router            = &backgroundRouter{}
	_ modules.IntegratableModule = &backgroundRouter{}
)

type backgroundRouter struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
	// gossipSub is used for broadcast communication
	// (i.e. multiple, unidentified receivers)
	// TECHDEBT: investigate diff between randomSub and gossipSub
	gossipSub *pubsub.PubSub
	// topic similar to pubsub but received messages are filtered by a "topic" string.
	// Published messages are also given the respective topic before broadcast.
	topic *pubsub.Topic
	// subscription provides an interface to continuously read messages from.
	subscription *pubsub.Subscription
	kadDHT       *dht.IpfsDHT
	pstore       typesP2P.Peerstore
}

func NewBackgroundRouter(bus modules.Bus, cfg *utils.RouterConfig) (typesP2P.Router, error) {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx := context.TODO()

	networkLogger := logger.Global.CreateLoggerForModule("backgroundRouter")
	networkLogger.Info().Msg("Initializing background")

	// seed initial peerstore with current on-chain peer info (i.e. staked actors)
	pstore, err := cfg.PeerstoreProvider.GetStakedPeerstoreAtHeight(
		cfg.CurrentHeightProvider.CurrentHeight(),
	)
	if err != nil {
		return nil, err
	}

	// NOTE_TO_SELF: `pubsub.NewRandomSub` requires a `size` arg.
	gossipSub, err := pubsub.NewGossipSub(ctx, cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("creating gossip pubsub: %w", err)
	}

	dhtMode := dht.ModeAutoServer
	// NB: don't act as a bootstrap node in peer discovery in client debug mode
	if isClientDebugMode(bus) {
		dhtMode = dht.ModeClient
	}

	kadDHT, err := dht.New(ctx, cfg.Host, dht.Mode(dhtMode))
	if err != nil {
		return nil, fmt.Errorf("creating DHT: %w", err)
	}

	topic, err := gossipSub.Join(protocol.BackgroundTopicStr)
	if err != nil {
		return nil, fmt.Errorf("joining background topic: %w", err)
	}

	// INVESTIGATE: `WithBufferSize` `SubOpt`:
	// > WithBufferSize is a Subscribe option to customize the size of the subscribe
	// > output buffer. The default length is 32 but it can be configured to avoid
	// > dropping messages if the consumer is not reading fast enough.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p-pubsub#WithBufferSize)
	subscription, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("subscribing to background topic: %w", err)
	}

	rtr := &backgroundRouter{
		host:         cfg.Host,
		gossipSub:    gossipSub,
		kadDHT:       kadDHT,
		topic:        topic,
		subscription: subscription,
		logger:       networkLogger,
		pstore:       pstore,
	}

	return rtr, nil
}

func (rtr *backgroundRouter) Broadcast(data []byte) error {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	return rtr.topic.Publish(context.TODO(), data)
}

func (rtr *backgroundRouter) Send(data []byte, address cryptoPocket.Address) error {
	peer := rtr.pstore.GetPeer(address)
	if peer == nil {
		return fmt.Errorf("peer with address %s not in peerstore", address)
	}

	if err := utils.Libp2pSendToPeer(rtr.host, data, peer); err != nil {
		return err
	}
	return nil
}

func (rtr *backgroundRouter) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil // intentional passthrough
}

func (rtr *backgroundRouter) GetPeerstore() typesP2P.Peerstore {
	return rtr.pstore
}

func (rtr *backgroundRouter) AddPeer(peer typesP2P.Peer) error {
	// Noop if peer with the pokt address already exists in the peerstore.
	// TECHDEBT: add method(s) to update peers.
	if p := rtr.pstore.GetPeer(peer.GetAddress()); p != nil {
		return nil
	}

	if err := utils.AddPeerToLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	return rtr.pstore.AddPeer(peer)
}

func (rtr *backgroundRouter) RemovePeer(peer typesP2P.Peer) error {
	if err := utils.RemovePeerFromLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	return rtr.pstore.RemovePeer(peer.GetAddress())
}

// isClientDebugMode returns the value of `ClientDebugMode` in the base config
func isClientDebugMode(bus modules.Bus) bool {
	return bus.GetRuntimeMgr().GetConfig().ClientDebugMode
}
