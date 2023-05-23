// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package background

import (
	"context"
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/config"
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

// backgroundRouter implements `typesP2P.Router` for use with all P2P participants.
type backgroundRouter struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	// handler is the function to call when a message is received.
	handler typesP2P.RouterHandler
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
	// gossipSub is used for broadcast communication
	// (i.e. multiple, unidentified receivers)
	// TECHDEBT: investigate diff between randomSub and gossipSub
	gossipSub *pubsub.PubSub
	// topic is similar to pubsub but received messages are filtered by a "topic" string.
	// Published messages are also given the respective topic before broadcast.
	topic *pubsub.Topic
	// subscription provides an interface to continuously read messages from.
	subscription *pubsub.Subscription
	// kadDHT is a kademlia distributed hash table used for routing and peer discovery.
	kadDHT *dht.IpfsDHT
	// TECHDEBT: `pstore` will likely be removed in future refactoring / simplification
	// of the `Router` interface.
	// pstore is the background router's peerstore.
	pstore typesP2P.Peerstore
}

// NewBackgroundRouter returns a `backgroundRouter` as a `typesP2P.Router`
// interface using the given configuration.
func NewBackgroundRouter(bus modules.Bus, cfg *config.BackgroundConfig) (typesP2P.Router, error) {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx := context.TODO()

	networkLogger := logger.Global.CreateLoggerForModule("backgroundRouter")
	networkLogger.Info().Msg("Initializing background router")

	// seed initial peerstore with current on-chain peer info (i.e. staked actors)
	pstore, err := cfg.PeerstoreProvider.GetStakedPeerstoreAtHeight(
		cfg.CurrentHeightProvider.CurrentHeight(),
	)
	if err != nil {
		return nil, err
	}

	// CONSIDERATION: If switching to `NewRandomSub`, there will be a max size
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

	go rtr.readSubscription(ctx)

	return rtr, nil
}

// Broadcast implements the respective `typesP2P.Router` interface  method.
func (rtr *backgroundRouter) Broadcast(data []byte) error {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	return rtr.topic.Publish(context.TODO(), data)
}

// Send implements the respective `typesP2P.Router` interface  method.
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

// GetPeerstore implements the respective `typesP2P.Router` interface  method.
func (rtr *backgroundRouter) GetPeerstore() typesP2P.Peerstore {
	return rtr.pstore
}

// AddPeer implements the respective `typesP2P.Router` interface  method.
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

// RemovePeer implements the respective `typesP2P.Router` interface  method.
func (rtr *backgroundRouter) RemovePeer(peer typesP2P.Peer) error {
	if err := utils.RemovePeerFromLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	return rtr.pstore.RemovePeer(peer.GetAddress())
}

func (rtr *backgroundRouter) readSubscription(ctx context.Context) {
	for msg, err := rtr.subscription.Next(ctx); ctx.Err() == nil; {
		if err != nil {
			rtr.logger.Error().Err(err).
				Msg("error reading from background topic subscription")
			continue
		}

		// TECHDEBT/DISCUSS: telemetry
		if err := rtr.handleBackgroundMsg(msg.Data); err != nil {
			rtr.logger.Error().Err(err).Msg("error handling background message")
			continue
		}
	}
}

func (rtr *backgroundRouter) handleBackgroundMsg(data []byte) error {
	var backgroundMsg typesP2P.BackgroundMessage
	if err := proto.Unmarshal(data, &backgroundMsg); err != nil {
		return err
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(data, &networkMessage); err != nil {
		return err
	}
	return nil
}

// isClientDebugMode returns the value of `ClientDebugMode` in the base config
func isClientDebugMode(bus modules.Bus) bool {
	return bus.GetRuntimeMgr().GetConfig().ClientDebugMode
}
