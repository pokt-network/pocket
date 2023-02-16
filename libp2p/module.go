package libp2p

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/libp2p/identity"
	libp2p2Network "github.com/pokt-network/pocket/libp2p/network"
	"github.com/pokt-network/pocket/libp2p/protocol"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider/persistence"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.P2PModule = &libp2pModule{}

type libp2pModule struct {
	logger                *modules.Logger
	bus                   modules.Bus
	cfg                   *configs.P2PConfig
	addrBookProvider      providers.AddrBookProvider
	currentHeightProvider providers.CurrentHeightProvider
	identity              libp2p.Option
	listenAddrs           libp2p.Option
	host                  host.Host
	pubsub                *pubsub.PubSub
	topic                 *pubsub.Topic
	subscription          *pubsub.Subscription
	network               typesP2P.Network
}

var (
	// TODO: consider security exposure of and "safe minimum" for timeout.
	// TODO: parameterize and expose via config.
	// readStreamTimeout is the duration to wait for a read operation on a
	// stream to complete, after which the stream is closed ("timed out").
	readStreamTimeoutDuration = time.Second * 10
	// ErrModule wraps errors which occur within the libp2pModule implementation.
	ErrModule = typesP2P.NewErrFactory("LibP2P module error")
)

func Create(bus modules.Bus) (modules.Module, error) {
	return CreateWithProviders(
		bus,
		persistence.NewPersistenceAddrBookProvider(bus),
		bus.GetConsensusModule(),
	)
}

func CreateWithProviders(
	bus modules.Bus,
	addrBookProvider addrbook_provider.AddrBookProvider,
	currentHeightProvider providers.CurrentHeightProvider,
) (modules.Module, error) {
	return new(libp2pModule).CreateWithProviders(
		bus,
		addrBookProvider,
		currentHeightProvider,
	)
}

func (mod *libp2pModule) GetModuleName() string {
	// TODO: double check if this should change.
	return modules.P2PModuleName
}

func (mod *libp2pModule) Create(bus modules.Bus) (modules.Module, error) {
	return Create(bus)
}

func (mod *libp2pModule) CreateWithProviders(
	bus modules.Bus,
	addrBookProvider addrbook_provider.AddrBookProvider,
	currentHeightProvider providers.CurrentHeightProvider,
) (modules.Module, error) {
	*mod = libp2pModule{
		logger:                new(modules.Logger),
		addrBookProvider:      addrBookProvider,
		currentHeightProvider: currentHeightProvider,
	}

	if err := bus.RegisterModule(mod); err != nil {
		return nil, ErrModule("unable to register module", err)
	}

	mod.logger.Print("Creating libp2p-backed network module")

	mod.cfg = bus.GetRuntimeMgr().GetConfig().P2P

	// INCOMPLETE: support RainTree network
	if mod.cfg.UseRainTree {
		return nil, ErrModule("raintree is not yet compatible with libp2p", nil)
	}

	// TECHDEBT: investigate any unnecessary
	// key exposure / duplication in memory
	secretKey, err := poktCrypto.NewLibP2PPrivateKey(mod.cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	mod.identity = libp2p.Identity(secretKey)

	switch mod.cfg.ConnectionType {
	case types.ConnectionType_TCPConnection:
		addr, err := mod.getMultiaddr()
		if err != nil {
			return nil, ErrModule("parsing multiaddr fom config", err)
		}
		mod.listenAddrs = libp2p.ListenAddrs(addr)
	case types.ConnectionType_EmptyConnection:
		mod.listenAddrs = nil
	default:
		return nil, ErrModule("", fmt.Errorf(
			// DISCUSS: should we refer to this as transport instead?
			"unsupported connection type: %s", mod.cfg.ConnectionType,
		))
	}

	if err := bus.RegisterModule(mod); err != nil {
		return nil, ErrModule("registering module", err)
	}
	return mod, nil
}

func (mod *libp2pModule) Start() error {
	// DISCUSS / CONSIDERATION: the linter fails with `hugeParam` when using
	// a value instead of a pointer. Do we want to change this everywhere?
	*mod.logger = logger.Global.CreateLoggerForModule("P2P")

	// IMPROVE: receive context in interface methods?
	ctx := context.Background()

	// TECHDEBT: metrics integration.
	var err error
	opts := []libp2p.Option{
		mod.identity,
		mod.listenAddrs,
		// TECHDEBT / INCOMPLETE: add transport security!
	}

	// NB: disable unused libp2p relay and ping services in client debug mode.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#DisableRelay
	// and https://pkg.go.dev/github.com/libp2p/go-libp2p#Ping)
	if !mod.GetBus().GetRuntimeMgr().GetConfig().ClientDebugMode {
		opts = append(opts,
			libp2p.DisableRelay(),
			libp2p.Ping(false),
		)
	}

	mod.host, err = libp2p.New(opts...)
	if err != nil {
		return ErrModule("unable to create libp2p host", err)
	}

	mod.logger.Info().Msgf("Listening on %s...", host.InfoFromHost(mod.host).Addrs)

	// TECHDEBT: use RandomSub or GossipSub once we're on more stable ground.
	// IMPROVE: consider supporting multiple router types via config.
	mod.pubsub, err = pubsub.NewFloodSub(ctx, mod.host)
	if err != nil {
		return ErrModule("unable to create pubsub", err)
	}

	// Topic is used to `#Publish` messages.
	mod.topic, err = mod.pubsub.Join(protocol.DefaultTopicStr)
	if err != nil {
		return ErrModule("unable to join pubsub topic", err)
	}

	// Subscription is notified when a new message is received on the topic.
	mod.subscription, err = mod.topic.Subscribe()
	if err != nil {
		return ErrModule("subscribing to pubsub topic", err)
	}

	mod.network, err = libp2p2Network.NewLibp2pNetwork(mod.bus, mod.addrBookProvider, mod.currentHeightProvider, mod.logger, mod.host, mod.topic)
	if err != nil {
		return ErrModule("creating network", err)
	}

	// NB: don't handle streams or read from the subscription in client debug mode.
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

func (mod *libp2pModule) Send(addr poktCrypto.Address, msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}

	return mod.network.NetworkSend(data, addr)
}

func (mod *libp2pModule) GetAddress() (poktCrypto.Address, error) {
	secretKey, err := poktCrypto.NewPrivateKey(mod.cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	return secretKey.Address(), nil
}

func (mod *libp2pModule) SetBus(bus modules.Bus) {
	// INVESTIGATE: Can the code flow be modified to set the bus here?
	// m.network.SetBus(m.GetBus())
	mod.bus = bus
}

func (mod *libp2pModule) GetBus() modules.Bus {
	if mod.bus == nil {
		mod.logger.Warn().Msg("PocketBus is not initialized")
		return nil
	}
	return mod.bus
}

// handleStream is called each time a peer establishes a new stream with this
// module's libp2p `host.Host`.
func (mod *libp2pModule) handleStream(stream network.Stream) {
	poktPeer, err := identity.PoktPeerFromStream(stream)
	if err != nil {
		mod.logger.Error().Err(err).Msgf("parsing remote peer public key, address: %s", poktPeer.Address)

		if err = stream.Close(); err != nil {
			mod.logger.Error().Err(err)
		}
	}

	if err := mod.network.AddPeerToAddrBook(poktPeer); err != nil {
		mod.logger.Error().Err(err).Msgf("adding remote peer to address book, address: %s", poktPeer.Address)
	}

	go mod.readStream(stream)
}

// readStream is intended to be called in a goroutine. It continuously reads from
// the given stream for handling at the network level. Used for handling "direct"
// messages (i.e. one specific target node).
func (mod *libp2pModule) readStream(stream network.Stream) {
	// NB: time out if no data is sent to free resources.
	if err := stream.SetReadDeadline(newReadStreamDeadline()); err != nil {
		mod.logger.Error().Err(err).Msg("setting stream read deadline")
		// TODO: abort if we can't set a read deadline?
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		mod.logger.Error().Err(err).Msg("reading from stream")
		if err := stream.Close(); err != nil {
			mod.logger.Error().Err(err)
		}
		// NB: abort this goroutine
		// TODO: signal this somewhere?
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			mod.logger.Error().Err(err)
		}
	}()

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

func (mod *libp2pModule) getMultiaddr() (multiaddr.Multiaddr, error) {
	// TECHDEBT: as soon as we add support for multiple transports
	// (i.e. not just TCP), we'll need to do something else.
	return identity.PeerMultiAddrFromServiceURL(fmt.Sprintf(
		"%s:%d", mod.cfg.Hostname, mod.cfg.ConsensusPort,
	))
}

// newReadStreamDeadline returns a future deadline
// based on the read stream timeout duration.
func newReadStreamDeadline() time.Time {
	return time.Now().Add(readStreamTimeoutDuration)
}
