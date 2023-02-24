/*
TECHDEBT: This module currently imports types  from the "legacy" P2P module.

Migration path:
 1. Redefine P2P concrete types in terms of interfaces
    - PeersManager (raintree/peersManager)
    - Peer (p2p/types/NetworkPeer)
    - AddrBook (p2p/types/AddrBook)
    - AddrBookMap (p2p/types/NetworkPeer)
    - rainTreeNetwork doesn't depend on any concrete p2p types
 2. Simplify libp2p module implementation
    - Transport likely reduces to nothing
    - Network interface can be simplified
    - Consider renaming network as it functions more like a "router"
    (NB: could be replaced in future iterations with a "raintree pubsub router")
 3. Remove "legacy" P2P module & rename libp2p module directory (possibly object names as well)
    - P2PModule interface can be simplified
    - Clean up TECHDEBT introduced in debug CLI and node startup
*/
package libp2p

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/libp2p/network"
	"github.com/pokt-network/pocket/libp2p/protocol"
	"github.com/pokt-network/pocket/logger"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.P2PModule = &libp2pModule{}

type libp2pModule struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	//nolint:unused // bus is used by embedded base module(s)
	bus         modules.Bus
	cfg         *configs.P2PConfig
	identity    libp2p.Option
	listenAddrs libp2p.Option
	// host encapsulates libp2p peerstore & connection manager
	host host.Host
	// pubsub is used for broadcast communication
	// (i.e. multiple, unidentified receivers)
	pubsub *pubsub.PubSub
	// topic similar to pubsub but received messages are filtered by a "topic" string.
	// Published messages are also given the respective topic before broadcast.
	topic *pubsub.Topic
	// subscription provides an interface to continuously read messages from.
	subscription *pubsub.Subscription
	network      typesP2P.Network
}

var (
	// TECHDEBT: configure timeouts. Consider security exposure vs. real-world conditions).
	// TECHDEBT: parameterize and expose via config.
	// readStreamTimeout is the duration to wait for a read operation on a
	// stream to complete, after which the stream is closed ("timed out").
	readStreamTimeoutDuration = time.Second * 10
)

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(libp2pModule).Create(bus, options...)
}

func (mod *libp2pModule) GetModuleName() string {
	return modules.P2PModuleName
}

func (mod *libp2pModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	logger.Global.Print("Creating libp2p-backed network module")

	*mod = libp2pModule{
		cfg: bus.GetRuntimeMgr().GetConfig().P2P,
	}

	// MUST call before referencing mod.bus to ensure != nil.
	bus.RegisterModule(mod)

	for _, option := range options {
		option(mod)
	}

	// TECHDEBT: investigate any unnecessary
	// key exposure / duplication in memory
	privateKey, err := crypto.NewLibP2PPrivateKey(mod.cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("loading private key: %w", err)
	}

	mod.identity = libp2p.Identity(privateKey)

	// INCOMPLETE: support RainTree network
	if mod.cfg.UseRainTree {
		return nil, fmt.Errorf("%s", "raintree is not yet compatible with libp2p")
	}

	switch mod.cfg.ConnectionType {
	case types.ConnectionType_TCPConnection:
		addr, err := mod.getMultiaddr()
		if err != nil {
			return nil, fmt.Errorf("parsing multiaddr from config: %w", err)
		}
		mod.listenAddrs = libp2p.ListenAddrs(addr)
	case types.ConnectionType_EmptyConnection:
		mod.listenAddrs = libp2p.NoListenAddrs
	default:
		return nil, fmt.Errorf(
			// DISCUSS: rename to "transport protocol" instead.
			"unsupported connection type: %s: %w",
			mod.cfg.ConnectionType,
			err,
		)
	}

	return mod, nil
}

func (mod *libp2pModule) Start() error {
	mod.logger = logger.Global.CreateLoggerForModule(mod.GetModuleName())

	// IMPROVE: receive context in interface methods.
	ctx := context.Background()

	// TECHDEBT: metrics integration.
	var err error
	opts := []libp2p.Option{
		mod.identity,
		// INCOMPLETE(#544): add transport security!
	}

	// Disable unused libp2p relay and ping services in client debug mode.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#DisableRelay
	// and https://pkg.go.dev/github.com/libp2p/go-libp2p#Ping)
	if mod.GetBus().GetRuntimeMgr().GetConfig().ClientDebugMode {
		opts = append(opts,
			libp2p.DisableRelay(),
			libp2p.Ping(false),
			libp2p.NoListenAddrs,
		)
	} else {
		opts = append(opts, mod.listenAddrs)
	}

	// Represents a libp2p network node, `libp2p.New` configures
	// and starts listening according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	mod.host, err = libp2p.New(opts...)
	if err != nil {
		return fmt.Errorf("unable to create libp2p host: %w", err)
	}

	listenAddrLogEvent := mod.logger.Info()
	for i, addr := range host.InfoFromHost(mod.host).Addrs {
		listenAddrLogEvent.Str(fmt.Sprintf("listen_addr_%d", i), addr.String())
	}
	listenAddrLogEvent.Msg("Listening for incoming connections...")

	// TECHDEBT: use RandomSub or GossipSub once we're on more stable ground.
	// IMPROVE: consider supporting multiple router types via config.
	mod.pubsub, err = pubsub.NewFloodSub(ctx, mod.host)
	if err != nil {
		return fmt.Errorf("unable to create pubsub: %w", err)
	}

	// Topic is used to `#Publish` messages.
	mod.topic, err = mod.pubsub.Join(protocol.DefaultTopicStr)
	if err != nil {
		return fmt.Errorf("unable to join pubsub topic: %w", err)
	}

	// Subscription is notified when a new message is received on the topic.
	mod.subscription, err = mod.topic.Subscribe()
	if err != nil {
		return fmt.Errorf("subscribing to pubsub topic: %w", err)
	}

	mod.network, err = network.NewLibp2pNetwork(mod.GetBus(), mod.logger, mod.host, mod.topic)
	if err != nil {
		return fmt.Errorf("creating network: %w", err)
	}

	// Don't handle streams or read from the subscription in client debug mode.
	if !mod.GetBus().GetRuntimeMgr().GetConfig().ClientDebugMode {
		mod.host.SetStreamHandler(protocol.PoktProtocolID, mod.handleStream)
		go mod.readFromSubscription(ctx)
	}
	return nil
}

func (mod *libp2pModule) Stop() error {
	return mod.host.Close()
}

func (mod *libp2pModule) Broadcast(msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}
	mod.logger.Info().Msg("broadcasting message to network")

	return mod.network.NetworkBroadcast(data)
}

func (mod *libp2pModule) Send(addr crypto.Address, msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}

	return mod.network.NetworkSend(data, addr)
}

func (mod *libp2pModule) GetAddress() (crypto.Address, error) {
	privateKey, err := crypto.NewPrivateKey(mod.cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	return privateKey.Address(), nil
}

// handleStream is called each time a peer establishes a new stream with this
// module's libp2p `host.Host`.
func (mod *libp2pModule) handleStream(stream libp2pNetwork.Stream) {
	peer, err := network.PeerFromLibp2pStream(stream)
	if err != nil {
		mod.logger.Error().Err(err).
			Str("address", peer.Address.String()).
			Msg("parsing remote peer public key")

		if err = stream.Close(); err != nil {
			mod.logger.Error().Err(err)
		}
	}

	if err := mod.network.AddPeerToAddrBook(peer); err != nil {
		mod.logger.Error().Err(err).
			Str("address", peer.Address.String()).
			Msg("adding remote peer to address book")
	}

	go mod.readStream(stream)
}

// readStream is intended to be called in a goroutine. It continuously reads from
// the given stream for handling at the network level. Used for handling "direct"
// messages (i.e. one specific target node).
func (mod *libp2pModule) readStream(stream libp2pNetwork.Stream) {
	closeStream := func() {
		if err := stream.Close(); err != nil {
			mod.logger.Error().Err(err)
		}
	}

	// NB: time out if no data is sent to free resources.
	if err := stream.SetReadDeadline(newReadStreamDeadline()); err != nil {
		mod.logger.Error().Err(err).Msg("setting stream read deadline")
		// TODO: abort if we can't set a read deadline?
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		mod.logger.Error().Err(err).Msg("reading from stream")
		closeStream()
		// NB: abort this goroutine
		// TODO: signal this somewhere?
		return
	}
	defer closeStream()

	mod.handleNetworkData(data)
}

// readFromSubscription is intended to be called in a goroutine. It continuously
// reads from the subscribed topic in preparation for handling at the network level.
// Used for handling "broadcast" messages (i.e. no specific target node).
func (mod *libp2pModule) readFromSubscription(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := mod.subscription.Next(ctx)
			if err != nil {
				mod.logger.Error().Err(err).Msg("reading from subscription")
			}

			// NB: ignore messages from self
			if msg.ReceivedFrom == mod.host.ID() {
				continue
			}

			mod.handleNetworkData(msg.Data)
		}
	}
}

func (mod *libp2pModule) handleNetworkData(data []byte) {
	appMsgData, err := mod.network.HandleNetworkData(data)
	if err != nil {
		mod.logger.Error().Err(err).Msg("handling network data")
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		return
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		mod.logger.Error().Err(err).Msg("Error decoding network message")
		return
	}

	event := messaging.PocketEnvelope{
		Content: networkMessage.Content,
	}

	mod.GetBus().PublishEventToBus(&event)
}

// getMultiaddr returns a multiaddr constructed from the `hostname` and `port`
// in the P2P config which pas provided upon creation.
func (mod *libp2pModule) getMultiaddr() (multiaddr.Multiaddr, error) {
	// TECHDEBT: as soon as we add support for multiple transports
	// (i.e. not just TCP), we'll need to do something else.
	return network.Libp2pMultiaddrFromServiceUrl(fmt.Sprintf(
		"%s:%d", mod.cfg.Hostname, mod.cfg.Port,
	))
}

// newReadStreamDeadline returns a future deadline
// based on the read stream timeout duration.
func newReadStreamDeadline() time.Time {
	return time.Now().Add(readStreamTimeoutDuration)
}

func (mod *libp2pModule) HandleEvent(msg *anypb.Any) error {
	return nil
}
