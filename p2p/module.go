package p2p

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
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

// TECHDEBT(#629): configure timeouts. Consider security exposure vs. real-world conditions.
// TECHDEBT(#629): parameterize and expose via config.
// readStreamTimeout is the duration to wait for a read operation on a
// stream to complete, after which the stream is closed ("timed out").
const readStreamTimeout = time.Second * 10

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	base_modules.IntegratableModule

	address        cryptoPocket.Address
	logger         *modules.Logger
	options        []modules.ModuleOption
	cfg            *configs.P2PConfig
	bootstrapNodes []string
	identity       libp2p.Option
	listenAddrs    libp2p.Option

	// Assigned during creation via `#setupDependencies()`.
	currentHeightProvider providers.CurrentHeightProvider
	pstoreProvider        providers.PeerstoreProvider

	// Assigned during `#Start()`. TLDR; `host` listens on instantiation.
	// and `network` depends on `host`.
	network typesP2P.Network
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options. Assigned via `#Start()` (starts on instantiation).
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
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

	m.options = options
	for _, option := range m.options {
		option(m)
	}

	if err := m.setupDependencies(); err != nil {
		return nil, err
	}

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

// Start instantiates and assigns `m.host`, unless one already exists, and
// `m.network` (which depends on `m.host` as a required config field).
func (m *p2pModule) Start() (err error) {
	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME,
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION,
		)

	// TECHDEBT: reconsider if this is acceptable as more `modules.ModuleOption`s
	// become supported. At time of writing, `WithHost()` is the only option
	// and it is only used in tests.
	// Re-evaluate options in case there is a `WithHost` option which would
	// assign`m.host`.
	for _, option := range m.options {
		option(m)
	}

	// Return early if host has already been started (e.g. via `WithHostOption`)
	if m.host == nil {
		if err = m.setupHost(); err != nil {
			return fmt.Errorf("setting up libp2pHost: %w", err)
		}
	}

	if err := m.setupNetwork(); err != nil {
		return fmt.Errorf("setting up network: %w", err)
	}

	// Don't handle incoming streams in client debug mode.
	if !m.isClientDebugMode() {
		m.host.SetStreamHandler(protocol.PoktProtocolID, m.handleStream)
	}

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME)
	return nil
}

func (m *p2pModule) Stop() error {
	err := m.host.Close()

	// Don't reuse closed host, `#Start()` will re-create.
	m.host = nil
	return err
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

// setupDependencies sets up the module's current height and peerstore providers.
func (m *p2pModule) setupDependencies() error {
	if err := m.setupCurrentHeightProvider(); err != nil {
		return err
	}

	if err := m.setupPeerstoreProvider(); err != nil {
		return err
	}

	return nil
}

// setupPeerstoreProvider attempts to retrieve the peerstore provider from the
// bus, if one is registered, otherwise returns a new `persistencePeerstoreProvider`.
func (m *p2pModule) setupPeerstoreProvider() error {
	m.logger.Debug().Msg("setupPeerstoreProvider")
	pstoreProviderModule, err := m.GetBus().GetModulesRegistry().GetModule(peerstore_provider.ModuleName)
	if err != nil {
		m.logger.Debug().Msg("creating new persistence peerstore...")
		pstoreProviderModule = persABP.NewPersistencePeerstoreProvider(m.GetBus())
	} else if pstoreProviderModule != nil {
		m.logger.Debug().Msg("loaded persistence peerstore...")
	}

	var ok bool
	m.pstoreProvider, ok = pstoreProviderModule.(providers.PeerstoreProvider)
	if !ok {
		return fmt.Errorf("unknown peerstore provider type: %T", pstoreProviderModule)
	}
	return nil
}

// setupCurrentHeightProvider attempts to retrieve the current height provider
// from the bus registry, falls back to the consensus module if none is registered.
func (m *p2pModule) setupCurrentHeightProvider() error {
	m.logger.Debug().Msg("setupCurrentHeightProvider")
	currentHeightProviderModule, err := m.GetBus().GetModulesRegistry().GetModule(current_height_provider.ModuleName)
	if err != nil {
		currentHeightProviderModule = m.GetBus().GetConsensusModule()
	}

	if currentHeightProviderModule == nil {
		return errors.New("no current height provider or consensus module registered")
	}

	m.logger.Debug().Msg("loaded current height provider")

	var ok bool
	m.currentHeightProvider, ok = currentHeightProviderModule.(providers.CurrentHeightProvider)
	if !ok {
		return fmt.Errorf("unexpected current height provider type: %T", currentHeightProviderModule)
	}
	return nil
}

// setupNetwork instantiates the configured network implementation.
func (m *p2pModule) setupNetwork() (err error) {
	if m.cfg.UseRainTree {
		m.network, err = raintree.NewRainTreeNetwork(
			m.GetBus(),
			raintree.RainTreeConfig{
				Host:                  m.host,
				Addr:                  m.address,
				PeerstoreProvider:     m.pstoreProvider,
				CurrentHeightProvider: m.currentHeightProvider,
			},
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

// setupHost creates a new libp2p host and assignes it to `m.host`. Libp2p host
// starts listening upon instantiation.
func (m *p2pModule) setupHost() (err error) {
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

	// TECHDEBT(#609): use `StringArrayLogMarshaler` post test-utilities refactor.
	addrStrs := make(map[int]string)
	for i, addr := range libp2pHost.InfoFromHost(m.host).Addrs {
		addrStrs[i] = addr.String()
	}
	m.logger.Info().Fields(addrStrs).Msg("Listening for incoming connections...")
	return nil
}

// isClientDebugMode returns the value of `ClientDebugMode` in the base config
func (m *p2pModule) isClientDebugMode() bool {
	return m.GetBus().GetRuntimeMgr().GetConfig().ClientDebugMode
}

// handleStream is called each time a peer establishes a new stream with this
// module's libp2p `host.Host`.
func (m *p2pModule) handleStream(stream libp2pNetwork.Stream) {
	m.logger.Debug().Msg("handling incoming stream")
	peer, err := utils.PeerFromLibp2pStream(stream)
	if err != nil {
		m.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("parsing remote peer identity")

		if err = stream.Reset(); err != nil {
			m.logger.Error().Err(err).Msg("resetting stream")
		}
		return
	}

	if err := m.network.AddPeer(peer); err != nil {
		m.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("adding remote peer to network")
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

	// debug logging: stream scope stats
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.27.0/core/network#StreamScope)
	// TECHDEBT: `logger.Global` is not a `*module.Logger`
	_logger := m.logger.Level(zerolog.DebugLevel)
	if err := utils.LogScopeStatFactory(
		&_logger,
		"stream scope (read-side)",
	)(stream.Scope()); err != nil {
		m.logger.Debug().Err(err).Msg("logging stream scope stats")
	}
	// ---

	data, err := io.ReadAll(stream)
	if err != nil {
		m.logger.Error().Err(err).Msg("reading from stream")
		if err := stream.Reset(); err != nil {
			m.logger.Debug().Err(err).Msg("resetting stream (read-side)")
		}
		return
	}

	if err := stream.Reset(); err != nil {
		m.logger.Debug().Err(err).Msg("resetting stream (read-side)")
	}

	// debug logging
	remotePeer, err := utils.PeerFromLibp2pStream(stream)
	if err != nil {
		m.logger.Debug().Err(err).Msg("getting remote remotePeer")
	} else {
		utils.LogIncomingMsg(m.logger, m.cfg.Hostname, remotePeer)
	}
	// ---

	if err := m.handleNetworkData(data); err != nil {
		m.logger.Error().Err(err).Msg("handling network data")
	}
}

// handleNetworkData passes a network message to the configured
// `Network`implementation for routing.
func (m *p2pModule) handleNetworkData(data []byte) error {
	appMsgData, err := m.network.HandleNetworkData(data)
	if err != nil {
		return err
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		return nil
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		return fmt.Errorf("decoding network message: %w", err)
	}

	event := messaging.PocketEnvelope{
		Content: networkMessage.Content,
	}
	m.GetBus().PublishEventToBus(&event)
	return nil
}

// getMultiaddr returns a multiaddr constructed from the `hostname` and `port`
// in the P2P config which was provided upon creation.
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
	return time.Now().Add(readStreamTimeout)
}
