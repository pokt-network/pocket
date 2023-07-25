// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package background

import (
	"context"
	"fmt"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/config"
	"github.com/pokt-network/pocket/p2p/protocol"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/unicast"
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ typesP2P.Router         = &backgroundRouter{}
	_ backgroundRouterFactory = &backgroundRouter{}
)

// TECHDEBT: Make these values configurable
// TECHDEBT: Consider using an exponential backoff instead
const (
	connectMaxRetries   = 5
	connectRetryTimeout = time.Second * 2
)

type backgroundRouterFactory = modules.FactoryWithConfig[typesP2P.Router, *config.BackgroundConfig]

// backgroundRouter implements `typesP2P.Router` for use with all P2P participants.
type backgroundRouter struct {
	base_modules.IntegrableModule
	unicast.UnicastRouter

	logger *modules.Logger
	// handler is the function to call when a message is received.
	handler typesP2P.MessageHandler
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
	// cancelReadSubscription is the cancel function for the context which is
	// monitored in the `#readSubscription()` go routine. Call to terminate it.
	// only one read subscription exists per router at any point in time
	cancelReadSubscription context.CancelFunc

	// Fields below are assigned during creation via `#setupDependencies()`.

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
	// TECHDEBT(#747, #749): `pstore` will likely be removed in future refactoring / simplification
	// of the `Router` interface.
	// pstore is the background router's peerstore. Assigned in `backgroundRouter#setupPeerstore()`.
	pstore typesP2P.Peerstore
}

// Create returns a `backgroundRouter` as a `typesP2P.Router`
// interface using the given configuration.
func Create(bus modules.Bus, cfg *config.BackgroundConfig) (typesP2P.Router, error) {
	return new(backgroundRouter).Create(bus, cfg)
}

func (*backgroundRouter) Create(bus modules.Bus, cfg *config.BackgroundConfig) (typesP2P.Router, error) {
	bgRouterLogger := logger.Global.CreateLoggerForModule("backgroundRouter")

	if err := cfg.IsValid(); err != nil {
		return nil, err
	}

	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	ctx, cancel := context.WithCancel(context.TODO())

	rtr := &backgroundRouter{
		logger:                 bgRouterLogger,
		handler:                cfg.Handler,
		host:                   cfg.Host,
		cancelReadSubscription: cancel,
	}
	bus.RegisterModule(rtr)

	bgRouterLogger.Info().Fields(map[string]any{
		"host_id":                cfg.Host.ID(),
		"unicast_protocol_id":    protocol.BackgroundProtocolID,
		"broadcast_pubsub_topic": protocol.BackgroundTopicStr,
	}).Msg("initializing background router")

	if err := rtr.setupDependencies(ctx, cfg); err != nil {
		return nil, err
	}

	go rtr.readSubscription(ctx)

	return rtr, nil
}

func (rtr *backgroundRouter) Close() error {
	rtr.logger.Debug().Msg("closing background router")

	rtr.cancelReadSubscription()
	rtr.subscription.Cancel()

	var topicCloseErr error
	if err := rtr.topic.Close(); err != context.Canceled {
		topicCloseErr = err
	}

	return multierr.Append(
		topicCloseErr,
		rtr.kadDHT.Close(),
	)
}

// GetModuleName implements the respective `modules.Integrable` interface  method.
func (rtr *backgroundRouter) GetModuleName() string {
	return typesP2P.UnstakedActorRouterSubmoduleName
}

// Broadcast implements the respective `typesP2P.Router` interface  method.
func (rtr *backgroundRouter) Broadcast(pocketEnvelopeBz []byte) error {
	backgroundMsg := &typesP2P.BackgroundMessage{
		Data: pocketEnvelopeBz,
	}
	backgroundMsgBz, err := proto.Marshal(backgroundMsg)
	if err != nil {
		return err
	}

	// TECHDEBT(#595): add ctx to interface methods and propagate down.
	return rtr.topic.Publish(context.TODO(), backgroundMsgBz)
}

// Send implements the respective `typesP2P.Router` interface  method.
func (rtr *backgroundRouter) Send(pocketEnvelopeBz []byte, address cryptoPocket.Address) error {
	backgroundMessage := &typesP2P.BackgroundMessage{
		Data: pocketEnvelopeBz,
	}
	backgroundMessageBz, err := proto.Marshal(backgroundMessage)
	if err != nil {
		return fmt.Errorf("marshalling background message: %w", err)
	}

	peer := rtr.pstore.GetPeer(address)
	if peer == nil {
		return fmt.Errorf("peer with address %s not in peerstore", address)
	}

	if err := utils.Libp2pSendToPeer(
		rtr.host,
		protocol.BackgroundProtocolID,
		backgroundMessageBz,
		peer,
	); err != nil {
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

// setupUnicastRouter configures and assigns `rtr.UnicastRouter`.
func (rtr *backgroundRouter) setupUnicastRouter() error {
	unicastRouterCfg := config.UnicastRouterConfig{
		Logger:         rtr.logger,
		Host:           rtr.host,
		ProtocolID:     protocol.BackgroundProtocolID,
		MessageHandler: rtr.handleBackgroundMsg,
		PeerHandler:    rtr.AddPeer,
	}

	unicastRouter, err := unicast.Create(rtr.GetBus(), &unicastRouterCfg)
	if err != nil {
		return fmt.Errorf("setting up unicast router: %w", err)
	}

	rtr.UnicastRouter = *unicastRouter
	return nil
}

// TECHBEDT(#810,#811): remove unused `BackgroundConfig`
func (rtr *backgroundRouter) setupDependencies(ctx context.Context, _ *config.BackgroundConfig) error {
	// NB: The order in which the internal components are setup below is important
	if err := rtr.setupUnicastRouter(); err != nil {
		return err
	}

	if err := rtr.setupPeerDiscovery(ctx); err != nil {
		return fmt.Errorf("setting up peer discovery: %w", err)
	}

	if err := rtr.setupPubsub(ctx); err != nil {
		return fmt.Errorf("setting up pubsub: %w", err)
	}

	if err := rtr.setupTopic(); err != nil {
		return fmt.Errorf("setting up topic: %w", err)
	}

	if err := rtr.setupSubscription(); err != nil {
		return fmt.Errorf("setting up subscription: %w", err)
	}

	if err := rtr.setupPeerstore(ctx); err != nil {
		return fmt.Errorf("setting up peerstore: %w", err)
	}
	return nil
}

func (rtr *backgroundRouter) setupPeerstore(ctx context.Context) (err error) {
	// TECHDEBT(#810, #811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule
	pstoreProviderModule, err := rtr.GetBus().GetModulesRegistry().
		GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	if err != nil {
		return fmt.Errorf("retrieving peerstore provider: %w", err)
	}
	pstoreProvider, ok := pstoreProviderModule.(providers.PeerstoreProvider)
	if !ok {
		return fmt.Errorf("unexpected peerstore provider type: %T", pstoreProviderModule)
	}

	rtr.logger.Debug().Msg("setupCurrentHeightProvider")
	currentHeightProvider := rtr.GetBus().GetCurrentHeightProvider()

	// seed initial peerstore with current on-chain peer info (i.e. staked actors)
	rtr.pstore, err = pstoreProvider.GetStakedPeerstoreAtHeight(
		currentHeightProvider.CurrentHeight(),
	)
	if err != nil {
		return err
	}

	// TECHDEBT(#859): integrate with `p2pModule#bootstrap()`.
	rtr.bootstrap(ctx)
	return nil
}

// setupPeerDiscovery sets up the Kademlia Distributed Hash Table (DHT)
func (rtr *backgroundRouter) setupPeerDiscovery(ctx context.Context) (err error) {
	dhtMode := dht.ModeAutoServer
	// NB: don't act as a bootstrap node in peer discovery in client debug mode
	if isClientDebugMode(rtr.GetBus()) {
		dhtMode = dht.ModeClient
	}

	rtr.kadDHT, err = dht.New(ctx, rtr.host, dht.Mode(dhtMode))
	return err
}

// setupPubsub sets up a new gossip sub topic using libp2p
func (rtr *backgroundRouter) setupPubsub(ctx context.Context) (err error) {
	// TECHDEBT(#730): integrate libp2p tracing via `pubsub.WithEventTracer()`.

	// CONSIDERATION: If switching to `NewRandomSub`, there will be a max size
	rtr.gossipSub, err = pubsub.NewGossipSub(ctx, rtr.host)
	return err
}

func (rtr *backgroundRouter) setupTopic() (err error) {
	if err := rtr.gossipSub.RegisterTopicValidator(
		protocol.BackgroundTopicStr,
		rtr.topicValidator,
	); err != nil {
		return fmt.Errorf(
			"registering topic validator for topic: %q: %w",
			protocol.BackgroundTopicStr, err,
		)
	}

	if rtr.topic, err = rtr.gossipSub.Join(protocol.BackgroundTopicStr); err != nil {
		return fmt.Errorf(
			"joining background topic: %q: %w",
			protocol.BackgroundTopicStr, err,
		)
	}
	return nil
}

func (rtr *backgroundRouter) setupSubscription() (err error) {
	// INVESTIGATE: `WithBufferSize` `SubOpt`:
	// > WithBufferSize is a Subscribe option to customize the size of the subscribe
	// > output buffer. The default length is 32 but it can be configured to avoid
	// > dropping messages if the consumer is not reading fast enough.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p-pubsub#WithBufferSize)
	rtr.subscription, err = rtr.topic.Subscribe()
	return err
}

// TECHDEBT(#859): integrate with `p2pModule#bootstrap()`.
func (rtr *backgroundRouter) bootstrap(ctx context.Context) {
	// CONSIDERATION: add `GetPeers` method, which returns a map,
	// to the `PeerstoreProvider` interface to simplify this loop.
	peerList := rtr.pstore.GetPeerList()
	for _, peer := range peerList {
		if err := utils.AddPeerToLibp2pHost(rtr.host, peer); err != nil {
			rtr.logger.Error().Err(err).Msg("adding peer to libp2p host")
			continue
		}

		libp2pAddrInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
		if err != nil {
			rtr.logger.Error().Err(err).Msg("converting peer info")
			continue
		}

		// don't attempt to connect to self
		if rtr.host.ID() == libp2pAddrInfo.ID {
			rtr.logger.Debug().Msg("not bootstrapping against self")
			continue
		}

		rtr.logger.Debug().Fields(map[string]any{
			"peer_id":   libp2pAddrInfo.ID.String(),
			"peer_addr": libp2pAddrInfo.Addrs[0].String(),
			"num_peers": len(peerList) - 1, // -1 as includes self
		}).Msg("connecting to peer")
		if err := rtr.connectWithRetry(ctx, libp2pAddrInfo); err != nil {
			rtr.logger.Error().Err(err).Msg("connecting to bootstrap peer")
			continue
		}
	}
}

// connectWithRetry attempts to connect to the given peer, retrying up to connectMaxRetries times
// and waiting connectRetryTimeout between each attempt.
func (rtr *backgroundRouter) connectWithRetry(ctx context.Context, libp2pAddrInfo libp2pPeer.AddrInfo) error {
	var err error
	for i := 0; i < connectMaxRetries; i++ {
		err = rtr.host.Connect(ctx, libp2pAddrInfo)
		if err == nil {
			return nil
		}

		rtr.logger.Error().Msgf("Failed to connect (attempt %d), retrying in %v...\n", i+1, connectRetryTimeout)
		time.Sleep(connectRetryTimeout)
	}

	return fmt.Errorf("failed to connect after %d attempts, last error: %w", 5, err)
}

// topicValidator is used in conjunction with libp2p-pubsub's notion of "topic
// validaton". It is used for arbitrary and concurrent pre-propagation validation
// of messages.
//
// (see: https://github.com/libp2p/specs/tree/master/pubsub#topic-validation
// and https://pkg.go.dev/github.com/libp2p/go-libp2p-pubsub#PubSub.RegisterTopicValidator)
//
// Also note: https://pkg.go.dev/github.com/libp2p/go-libp2p-pubsub#BasicSeqnoValidator
func (rtr *backgroundRouter) topicValidator(_ context.Context, _ libp2pPeer.ID, msg *pubsub.Message) bool {
	var backgroundMsg typesP2P.BackgroundMessage
	if err := proto.Unmarshal(msg.Data, &backgroundMsg); err != nil {
		rtr.logger.Error().Err(err).Msg("unmarshalling Background message")
		return false
	}

	if backgroundMsg.Data == nil {
		rtr.logger.Debug().Msg("no data in Background message")
		return false
	}

	poktEnvelope := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(backgroundMsg.Data, &poktEnvelope); err != nil {
		rtr.logger.Error().Err(err).Msg("Error decoding Background message")
		return false
	}

	return true
}

// readSubscription is a while loop for receiving and handling messages from the
// subscription. It is intended to be called as a goroutine.
func (rtr *backgroundRouter) readSubscription(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			if err != context.Canceled {
				rtr.logger.Error().Err(err).
					Msg("context error while reading subscription")
			}
			return
		}
		msg, err := rtr.subscription.Next(ctx)
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

func (rtr *backgroundRouter) handleBackgroundMsg(backgroundMsgBz []byte) error {
	var backgroundMsg typesP2P.BackgroundMessage
	if err := proto.Unmarshal(backgroundMsgBz, &backgroundMsg); err != nil {
		return err
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if backgroundMsg.Data == nil {
		return nil
	}

	return rtr.handler(backgroundMsg.Data)
}

// isClientDebugMode returns the value of `ClientDebugMode` in the base config
func isClientDebugMode(bus modules.Bus) bool {
	return bus.GetRuntimeMgr().GetConfig().ClientDebugMode
}
