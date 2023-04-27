// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	"context"
	"fmt"
	libp2pDiscoveryUtil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/pokt-network/pocket/p2p/protocol"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	libp2pDiscovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ typesP2P.Network           = &router{}
	_ modules.IntegratableModule = &router{}
)

type router struct {
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
	logger       *modules.Logger
}

func NewNetwork(host libp2pHost.Host, pstoreProvider providers.PeerstoreProvider, currentHeightProvider providers.CurrentHeightProvider) (typesP2P.Network, error) {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx := context.TODO()

	networkLogger := logger.Global.CreateLoggerForModule("router")
	networkLogger.Info().Msg("Initializing stdnetwork")

	// seed initial peerstore with current on-chain peer info (i.e. staked actors)
	pstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		return nil, err
	}

	// NOTE_TO_SELF: `pubsub.NewRandomSub` requires a `size` arg.
	gossipSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("creating gossip pubsub: %w", err)
	}

	kadDHT, err := dht.New(ctx, host, dht.Mode(dht.ModeAutoServer))
	if err != nil {
		return nil, fmt.Errorf("creating DHT: %w", err)
	}

	// TODO_THIS_COMMIT: does this need to happen here? (or at all??)
	//if err = kadDHT.Bootstrap(ctx); err != nil {
	//	return nil, fmt.Errorf("bootstrapping DHT: %w", err)
	//}

	topic, err := gossipSub.Join(protocol.BackgroundTopicStr)
	if err != nil {
		return nil, fmt.Errorf("joining background topic: %w", err)
	}

	// TODO_THIS_COMMIT: check out what subscribe options exist
	subscription, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("subscribing to background topic: %w", err)
	}

	rtr := &router{
		host:         host,
		gossipSub:    gossipSub,
		kadDHT:       kadDHT,
		topic:        topic,
		subscription: subscription,
		logger:       networkLogger,
		pstore:       pstore,
	}

	// kick off peer discovery
	discovery := libp2pDiscovery.NewRoutingDiscovery(rtr.kadDHT)
	go rtr.discover(ctx, discovery)

	// TODO_THIS_COMMIT: check out what `discovery.Option`'s exist
	libp2pDiscoveryUtil.Advertise(ctx, discovery, protocol.PeerDiscoveryNamespace)

	return rtr, nil
}

func (rtr *router) NetworkBroadcast(data []byte) error {
	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	return rtr.topic.Publish(context.TODO(), data)
}

func (rtr *router) NetworkSend(data []byte, address cryptoPocket.Address) error {
	peer := rtr.pstore.GetPeer(address)
	if peer == nil {
		return fmt.Errorf("peer with address %s not in peerstore", address)
	}

	if err := utils.Libp2pSendToPeer(rtr.host, data, peer); err != nil {
		return err
	}
	return nil
}

func (rtr *router) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil // intentional passthrough
}

func (rtr *router) GetPeerstore() typesP2P.Peerstore {
	return rtr.pstore
}

func (rtr *router) AddPeer(peer typesP2P.Peer) error {
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

func (rtr *router) RemovePeer(peer typesP2P.Peer) error {
	if err := utils.RemovePeerFromLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	return rtr.pstore.RemovePeer(peer.GetAddress())
}

func (rtr *router) GetBus() modules.Bus  { return nil }
func (rtr *router) SetBus(_ modules.Bus) {}

// discover is intended to run in its own go routine. It
func (rtr *router) discover(ctx context.Context, discovery *libp2pDiscovery.RoutingDiscovery) {
	// TECHDEBT: parameterize
	//ticker := time.NewTicker(time.Second * 1)
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peerCh, err := discovery.FindPeers(ctx, protocol.PeerDiscoveryNamespace)
			//rtr.logger.Debug().Int("numPeers", len(peerCh)).Msg("")

			if err != nil {
				rtr.logger.Error().Err(err).Msg("finding peers")
				continue
			}
			// TECHDEBT: ensure self isn't considered in the count

			for peer := range peerCh {
				if peer.ID == rtr.host.ID() {
					// self-discovery: no-op
					continue
				}

				//rtr.logger.Debug().Str("ID", peer.ID.String()).Msg("discovered peer")
				//rtr.logger.Debug().Msgf("connectedness: %s", rtr.host.Network().Connectedness(peer.ID).String())
				// TODO_THIS_COMMIT: reconsider
				if rtr.host.Network().Connectedness(peer.ID) != libp2pNetwork.Connected {
					// TODO_THIS_COMMIT: why `#DIalPeer()` over `#Connect()`
					_, err = rtr.host.Network().DialPeer(ctx, peer.ID)
					rtr.logger.Debug().Str("peerID", peer.ID.String()).Msg("Connected to peer")

					if err != nil {
						rtr.logger.Error().Str("peerID", peer.ID.String()).Msg("dialing peer")
						continue
					}
				}
			}
		}
	}
}
