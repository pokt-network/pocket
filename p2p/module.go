package p2p

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/protocol"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	persABP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/persistence"
	"github.com/pokt-network/pocket/p2p/raintree"
	"github.com/pokt-network/pocket/p2p/stdnetwork"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/telemetry"
)

const readStreamTimeoutDuration = time.Second * 10

// TECHDEBT: configure timeouts. Consider security exposure vs. real-world conditions).
// TECHDEBT: parameterize and expose via config.
// readStreamTimeout is the duration to wait for a read operation on a
// stream to complete, after which the stream is closed ("timed out").
var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	base_modules.IntegratableModule

	address               cryptoPocket.Address
	logger                *modules.Logger
	cfg                   *configs.P2PConfig
	bootstrapNodes        []string
	currentHeightProvider providers.CurrentHeightProvider
	pstoreProvider        providers.PeerstoreProvider
	identity              libp2p.Option
	listenAddrs           libp2p.Option
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
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

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(p2pModule).Create(bus, options...)
}

// WithHostOption associates an existing (i.e. "started") libp2p `host.Host`
// with this module, instead of creating a new one on `#Start()`.
func WithHostOption(host libp2pHost.Host) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*p2pModule)
		if ok {
			mod.host = host
			mod.logger.Debug().Msg("using host provided via `WithHostOption`")
		}
	}
}

func (m *p2pModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	logger.Global.Debug().Msg("Creating libp2p-backed network module")
	*m = p2pModule{
		cfg:    bus.GetRuntimeMgr().GetConfig().P2P,
		logger: logger.Global.CreateLoggerForModule(modules.P2PModuleName),
	}

	// MUST call before referencing m.bus to ensure != nil.
	bus.RegisterModule(m)
	m.setupDependencies()

	for _, option := range options {
		option(m)
	}

	if err := m.configureBootstrapNodes(); err != nil {
		return nil, err
	}

	// TECHDEBT: investigate any unnecessary
	// key exposure / duplication in memory
	privateKey, err := cryptoPocket.NewPrivateKey(m.cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parsing private key as pocket key: %w", err)
	}
	m.address = privateKey.Address()

	libp2pPrivKey, err := cryptoPocket.NewLibP2PPrivateKey(m.cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parsing private key as libp2p key: %w", err)
	}
	m.identity = libp2p.Identity(libp2pPrivKey)

	switch m.cfg.ConnectionType {
	case types.ConnectionType_TCPConnection:
		addr, err := m.getMultiaddr()
		if err != nil {
			return nil, fmt.Errorf("parsing multiaddr from config: %w", err)
		}
		m.listenAddrs = libp2p.ListenAddrs(addr)
	case types.ConnectionType_EmptyConnection:
		m.listenAddrs = libp2p.NoListenAddrs
	default:
		return nil, fmt.Errorf(
			// TECHDEBT: rename to "transport protocol" instead.
			"unsupported connection type: %s: %w",
			m.cfg.ConnectionType,
			err,
		)
	}

	return m, nil
}
func (m *p2pModule) GetModuleName() string {
	return modules.P2PModuleName
}

func (m *p2pModule) Start() (err error) {
	// TECHDEBT(#595): receive context in interface methods.
	ctx := context.Background()

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME,
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION,
		)

	if err = m.startHost(); err != nil {
		return fmt.Errorf("starting libp2pHost: %w", err)
	}

	listenAddrLogEvent := m.logger.Info()
	for i, addr := range libp2pHost.InfoFromHost(m.host).Addrs {
		listenAddrLogEvent.Str(fmt.Sprintf("listen_addr_%d", i), addr.String())
	}
	listenAddrLogEvent.Msg("Listening for incoming connections...")

	// TECHDEBT: use RandomSub or GossipSub once we're on more stable ground.
	// IMPROVE: consider supporting multiple router types via config.
	m.pubsub, err = pubsub.NewFloodSub(ctx, m.host)
	if err != nil {
		return fmt.Errorf("unable to create pubsub: %w", err)
	}

	// Topic is used to `#Publish` messages.
	m.topic, err = m.pubsub.Join(protocol.DefaultTopicStr)
	if err != nil {
		return fmt.Errorf("unable to join pubsub topic: %w", err)
	}

	// Subscription is notified when a new message is received on the topic.
	m.subscription, err = m.topic.Subscribe()
	if err != nil {
		return fmt.Errorf("subscribing to pubsub topic: %w", err)
	}

	if err := m.startNetwork(); err != nil {
		return fmt.Errorf("creating network: %w", err)
	}

	// Don't handle streams or read from the subscription in client debug mode.
	if !m.isClientDebugMode() {
		m.host.SetStreamHandler(protocol.PoktProtocolID, m.handleStream)
		go m.readFromSubscription(ctx)
	}

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME)
	return nil
}

func (m *p2pModule) Stop() error {
	return m.host.Close()
}

func (m *p2pModule) Broadcast(msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	//TECHDEBT: use shared/codec for marshalling
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}
	m.logger.Info().Msg("broadcasting message to network")

	return m.network.NetworkBroadcast(data)
}

func (m *p2pModule) Send(addr cryptoPocket.Address, msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	//TECHDEBT: use shared/codec for marshalling
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}

	return m.network.NetworkSend(data, addr)
}

// TECHDEBT(#348): Define what the node identity is throughout the codebase
func (m *p2pModule) GetAddress() (cryptoPocket.Address, error) {
	return m.address, nil
}

func (m *p2pModule) setupDependencies() {
	m.setupCurrentHeightProvider()
	m.setupPeerstoreProvider()
}

// setupPeerstoreProvider attempts to retrieve the peerstore provider from the
// bus, if one is registered, otherwise returns a new `persistencePeerstoreProvider`.
func (m *p2pModule) setupPeerstoreProvider() {
	m.logger.Debug().Msg("setupPeerstoreProvider")
	pstoreProviderModule, err := m.GetBus().GetModulesRegistry().GetModule(peerstore_provider.ModuleName)
	if pstoreProviderModule != nil {
		m.logger.Debug().Msg("loaded persistence peerstore...")
	}
	if err != nil {
		m.logger.Debug().Msg("NewPersistencePeerstore...")
		pstoreProviderModule = persABP.NewPersistencePeerstoreProvider(m.GetBus())
	}

	var ok bool
	m.pstoreProvider, ok = pstoreProviderModule.(providers.PeerstoreProvider)
	if !ok {
		m.logger.Fatal().Msgf("unknown peerstore provider type: %T", pstoreProviderModule)
	}
}

// setupCurrentHeightProvider attempts to retrieve the current height provider
// from the bus registry, falls back to the consensus module if none is registered.
func (m *p2pModule) setupCurrentHeightProvider() {
	currentHeightProviderModule, err := m.GetBus().GetModulesRegistry().GetModule(current_height_provider.ModuleName)
	if err != nil {
		currentHeightProviderModule = m.GetBus().GetConsensusModule()
	}

	var ok bool
	m.currentHeightProvider, ok = currentHeightProviderModule.(providers.CurrentHeightProvider)
	if !ok {
		m.logger.Fatal().Msgf("unexpected current height provider type: %T", currentHeightProviderModule)
	}
}

func (m *p2pModule) startNetwork() (err error) {
	if m.cfg.UseRainTree {
		m.network = raintree.NewRainTreeNetwork(
			m.address,
			m.GetBus(),
			m.pstoreProvider,
			m.currentHeightProvider,
		)
	} else {
		m.network, err = stdnetwork.NewNetwork(
			m.host,
			m.pstoreProvider,
			m.currentHeightProvider,
		)
	}
	return err
}

func (m *p2pModule) startHost() (err error) {
	// Return early if host has already been started (e.g. via `WithHostOption`)
	if m.host != nil {
		return nil
	}
	opts := []libp2p.Option{
		// Explicitly specify supported transport security options (noise, TLS)
		// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.26.3#DefaultSecurity)
		libp2p.DefaultSecurity,
		m.identity,
	}

	// Disable unused libp2p relay and ping services in client debug mode.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#DisableRelay
	// and https://pkg.go.dev/github.com/libp2p/go-libp2p#Ping)
	if m.isClientDebugMode() {
		opts = append(opts,
			libp2p.DisableRelay(),
			libp2p.Ping(false),
			libp2p.NoListenAddrs,
		)
	} else {
		opts = append(opts, m.listenAddrs)
	}

	m.host, err = libp2p.New(opts...)
	if err != nil {
		return fmt.Errorf("unable to create libp2p host: %w", err)
	}
	return nil
}

func (m *p2pModule) isClientDebugMode() bool {
	return m.GetBus().GetRuntimeMgr().GetConfig().ClientDebugMode
}

// handleStream is called each time a peer establishes a new stream with this
// module's libp2p `host.Host`.
func (m *p2pModule) handleStream(stream libp2pNetwork.Stream) {
	peer, err := utils.PeerFromLibp2pStream(stream)
	if err != nil {
		m.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("parsing remote peer public key")

		if err = stream.Reset(); err != nil {
			m.logger.Error().Err(err).Msg("resetting stream")
		}
	}

	if err := m.network.AddPeer(peer); err != nil {
		m.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("adding remote peer to address book")
	}

	go m.readStream(stream)
}

// readStream is intended to be called in a goroutine. It continuously reads from
// the given stream for handling at the network level. Used for handling "direct"
// messages (i.e. one specific target node).
func (m *p2pModule) readStream(stream libp2pNetwork.Stream) {
	// Time out if no data is sent to free resources.
	if err := stream.SetReadDeadline(newReadStreamDeadline()); err != nil {
		// NB: tests using libp2p's `mocknet` rely on this not returning an error.
		// `SetReadDeadline` not supported by `mocknet` streams.
		m.logger.Debug().Err(err).Msg("setting stream read deadline")
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		m.logger.Error().Err(err).Msg("reading from stream")
		if err := stream.Reset(); err != nil {
			m.logger.Debug().Err(err).Msg("resetting stream")
		}
		return
	}

	m.handleNetworkData(data)

	if err := stream.CloseRead(); err != nil {
		m.logger.Debug().Err(err).Msg("closing read stream")
	}
}

// readFromSubscription is intended to be called in a goroutine. It continuously
// reads from the subscribed topic in preparation for handling at the network level.
// Used for handling "broadcast" messages (i.e. no specific target node).
func (m *p2pModule) readFromSubscription(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := m.subscription.Next(ctx)
			if err != nil {
				m.logger.Error().Err(err).
					Bool("TODO", true).
					Msg("reading from subscription")
			}

			// Ignore messages from self
			if msg.ReceivedFrom == m.host.ID() {
				continue
			}

			m.handleNetworkData(msg.Data)
		}
	}
}

func (m *p2pModule) handleNetworkData(data []byte) {
	appMsgData, err := m.network.HandleNetworkData(data)
	if err != nil {
		m.logger.Error().Err(err).Msg("handling network data")
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		return
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		m.logger.Error().Err(err).
			Bool("TODO", true).
			Msg("Error decoding network message")
		return
	}

	event := messaging.PocketEnvelope{
		Content: networkMessage.Content,
	}

	m.GetBus().PublishEventToBus(&event)
}

// getMultiaddr returns a multiaddr constructed from the `hostname` and `port`
// in the P2P config which pas provided upon creation.
func (m *p2pModule) getMultiaddr() (multiaddr.Multiaddr, error) {
	// TECHDEBT: as soon as we add support for multiple transports
	// (i.e. not just TCP), we'll need to do something else.
	return utils.Libp2pMultiaddrFromServiceURL(fmt.Sprintf(
		"%s:%d", m.cfg.Hostname, m.cfg.Port,
	))
}

// newReadStreamDeadline returns a future deadline
// based on the read stream timeout duration.
func newReadStreamDeadline() time.Time {
	return time.Now().Add(readStreamTimeoutDuration)
}
