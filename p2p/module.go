package p2p

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/libp2p/go-libp2p"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/background"
	"github.com/pokt-network/pocket/p2p/config"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	persPSP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/persistence"
	"github.com/pokt-network/pocket/p2p/raintree"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/telemetry"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	base_modules.IntegrableModule

	started        atomic.Bool
	address        cryptoPocket.Address
	logger         *modules.Logger
	options        []modules.ModuleOption
	cfg            *configs.P2PConfig
	bootstrapNodes []string
	identity       libp2p.Option
	listenAddrs    libp2p.Option

	// TECHDEBT(#810): register the providers to the module registry instead of
	// holding a reference in the module struct and passing via router config.
	//
	// Assigned during creation via `#setupDependencies()`.
	currentHeightProvider providers.CurrentHeightProvider
	pstoreProvider        providers.PeerstoreProvider
	nonceDeduper          *mempool.GenericFIFOSet[uint64, uint64]

	// TECHDEBT(#810): register the routers to the module registry instead of
	// holding a reference in the module struct. This will improve testability.
	//
	// Assigned during `#Start()`. TLDR; `host` listens on instantiation.
	// `stakedActorRouter` and `unstakedActorRouter` depends on `host`.
	stakedActorRouter   typesP2P.Router
	unstakedActorRouter typesP2P.Router
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options. Assigned via `#Start()` (starts on instantiation).
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(p2pModule).Create(bus, options...)
}

func (m *p2pModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	logger.Global.Debug().Msg("Creating P2P module")
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
// `m.stakedActorRouter` (which depends on `m.host` as a required config field).
func (m *p2pModule) Start() (err error) {
	if !m.started.CompareAndSwap(false, true) {
		return fmt.Errorf("p2p module already started")
	}

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME,
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION,
		)

	// Return early if host has already been started (e.g. via `WithHost`)
	if m.host == nil {
		// Libp2p hosts provided via `WithHost()` option are destroyed when
		// `#Stop()`ing the module. Therefore, a new one must be created.
		// The new host may be configured differently than that which was
		// provided originally in `WithHost()`.
		if len(m.options) != 0 {
			m.logger.Warn().Msg("creating new libp2p host")
		}

		if err = m.setupHost(); err != nil {
			return fmt.Errorf("setting up libp2pHost: %w", err)
		}
	}

	if err := m.setupRouters(); err != nil {
		return fmt.Errorf("setting up routers: %w", err)
	}

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME)
	return nil
}

func (m *p2pModule) Stop() error {
	m.logger.Debug().Msg("stopping P2P module")

	if !m.started.CompareAndSwap(true, false) {
		return fmt.Errorf("p2p module already stopped")
	}

	var stakedActorRouterCloseErr error
	if m.stakedActorRouter != nil {
		stakedActorRouterCloseErr = m.stakedActorRouter.Close()
	}

	routerCloseErrs := multierr.Append(
		m.unstakedActorRouter.Close(),
		stakedActorRouterCloseErr,
	)

	err := multierr.Append(
		routerCloseErrs,
		m.host.Close(),
	)

	// Don't reuse closed host, `#Start()` will re-create.
	m.host = nil
	m.stakedActorRouter = nil
	m.unstakedActorRouter = nil
	return err
}

func (m *p2pModule) Broadcast(msg *anypb.Any) error {
	isStaked, err := m.isStakedActor()
	if err != nil {
		return err
	}

	if isStaked {
		if m.stakedActorRouter == nil {
			return fmt.Errorf("broadcasting: staked actor router not started")
		}
	}

	if m.unstakedActorRouter == nil {
		return fmt.Errorf("broadcasting: unstaked actor router not started")
	}

	poktEnvelope := &messaging.PocketEnvelope{
		Content: msg,
		Nonce:   cryptoPocket.GetNonce(),
	}
	poktEnvelopeBz, err := codec.GetCodec().Marshal(poktEnvelope)
	if err != nil {
		return err
	}

	var stakedBroadcastErr error
	if isStaked {
		stakedBroadcastErr = m.stakedActorRouter.Broadcast(poktEnvelopeBz)
	}

	unstakedBroadcastErr := m.unstakedActorRouter.Broadcast(poktEnvelopeBz)

	return multierr.Append(stakedBroadcastErr, unstakedBroadcastErr)
}

func (m *p2pModule) Send(addr cryptoPocket.Address, msg *anypb.Any) error {
	poktEnvelope := &messaging.PocketEnvelope{
		Content: msg,
		Nonce:   cryptoPocket.GetNonce(),
	}

	poktEnvelopeBz, err := codec.GetCodec().Marshal(poktEnvelope)
	if err != nil {
		return err
	}

	isStaked, err := m.isStakedActor()
	if err != nil {
		return err
	}

	// Send via the staked actor router both if this node and the peer are staked
	// actors; otherwise, send via the unstaked actor router.
	if !isStaked {
		return m.unstakedActorRouter.Send(poktEnvelopeBz, addr)
	}

	stakedActorSendErr := m.stakedActorRouter.Send(poktEnvelopeBz, addr)

	// Peer is not a staked actor.
	if errors.Is(stakedActorSendErr, typesP2P.ErrUnknownPeer) {
		m.logger.Warn().
			Str("address", addr.String()).
			Msgf("attempting to send to unstaked actor")

		return m.unstakedActorRouter.Send(poktEnvelopeBz, addr)
	}
	return nil
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

	if err := m.setupNonceDeduper(); err != nil {
		return err
	}
	return nil
}

// setupPeerstoreProvider attempts to retrieve the peerstore provider from the
// bus, if one is registered, otherwise returns a new `persistencePeerstoreProvider`.
func (m *p2pModule) setupPeerstoreProvider() error {
	m.logger.Debug().Msg("setupPeerstoreProvider")

	// TECHDEBT(#810, #811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule
	pstoreProviderModule, err := m.GetBus().
		GetModulesRegistry().
		GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	if err != nil {
		m.logger.Debug().Msg("creating new persistence peerstore...")
		pstoreProvider, err := persPSP.Create(m.GetBus())
		if err != nil {
			return err
		}

		m.pstoreProvider = pstoreProvider
		return nil
	}

	m.logger.Debug().Msg("loaded persistence peerstore...")
	pstoreProvider, ok := pstoreProviderModule.(providers.PeerstoreProvider)
	if !ok {
		return fmt.Errorf("unknown peerstore provider type: %T", pstoreProviderModule)
	}

	// TECHDEBT(#810): register the provider to the module registry instead of
	// holding a reference in the module struct and passing via router config.
	m.pstoreProvider = pstoreProvider

	return nil
}

// setupCurrentHeightProvider attempts to retrieve the current height provider
// from the bus registry, falls back to the consensus module if none is registered.
func (m *p2pModule) setupCurrentHeightProvider() error {
	// TECHDEBT(#810): simplify once submodules are more convenient to retrieve.
	m.logger.Debug().Msg("setupCurrentHeightProvider")
	currentHeightProviderModule, err := m.GetBus().GetModulesRegistry().GetModule(current_height_provider.ModuleName)
	if err != nil {
		// TECHDEBT(#810): add a `consensusCurrentHeightProvider` submodule to wrap
		// the consensus module usage (similar to how `persistencePeerstoreProvider`
		// wraps persistence).
		currentHeightProviderModule = m.GetBus().GetConsensusModule()
	}

	if currentHeightProviderModule == nil {
		return errors.New("no current height provider or consensus module registered")
	}

	m.logger.Debug().Msg("loaded current height provider")

	currentHeightProvider, ok := currentHeightProviderModule.(providers.CurrentHeightProvider)
	if !ok {
		return fmt.Errorf("unexpected current height provider type: %T", currentHeightProviderModule)
	}

	// TECHDEBT(#810): register the provider to the module registry instead of
	// holding a reference in the module struct and passing via router config.
	m.currentHeightProvider = currentHeightProvider

	return nil
}

// setupNonceDeduper initializes an empty deduper with a max capacity of
// the configured `MaxNonces`.
func (m *p2pModule) setupNonceDeduper() error {
	if m.cfg.MaxNonces == 0 {
		return fmt.Errorf("max nonces must be greater than 0")
	}

	m.nonceDeduper = utils.NewNonceDeduper(m.cfg.MaxNonces)
	return nil
}

// setupRouters instantiates the configured router implementations.
func (m *p2pModule) setupRouters() (err error) {
	// TECHDEBT(#810): register the routers to the module registry instead of
	// holding a reference in the module struct. This will improve testability.
	if err := m.setupStakedRouter(); err != nil {
		return err
	}

	if err := m.setupUnstakedRouter(); err != nil {
		return err
	}
	return nil
}

// setupStakedRouter initializes the staked actor router ONLY IF this node is
// a staked actor, exclusively for use between staked actors.
func (m *p2pModule) setupStakedRouter() (err error) {
	// `nstakedActorRouter` may already be initialized via a `ModuleOption`.
	if m.stakedActorRouter != nil {
		m.logger.Debug().Msg("staked actor router already initialized")
		return nil
	}

	if isStaked, err := m.isStakedActor(); err != nil {
		return err
	} else if !isStaked {
		return nil
	}

	m.logger.Debug().Msg("setting up staked actor router")
	m.stakedActorRouter, err = raintree.NewRainTreeRouter(
		m.GetBus(),
		&config.RainTreeConfig{
			Addr:                  m.address,
			CurrentHeightProvider: m.currentHeightProvider,
			PeerstoreProvider:     m.pstoreProvider,
			Host:                  m.host,
			Handler:               m.handlePocketEnvelope,
		},
	)
	if err != nil {
		return fmt.Errorf("setting up staked actor router: %w", err)
	}
	return nil
}

// setupUnstakedRouter initializes the unstaked actor router for use with the
// entire P2P network.
func (m *p2pModule) setupUnstakedRouter() (err error) {
	// `unstakedActorRouter` may already be initialized via a `ModuleOption`.
	if m.unstakedActorRouter != nil {
		m.logger.Debug().Msg("unstaked actor router already initialized")
		return nil
	}

	m.logger.Debug().Msg("setting up unstaked actor router")
	m.unstakedActorRouter, err = background.Create(
		m.GetBus(),
		&config.BackgroundConfig{
			Addr:                  m.address,
			CurrentHeightProvider: m.currentHeightProvider,
			PeerstoreProvider:     m.pstoreProvider,
			Host:                  m.host,
			Handler:               m.handlePocketEnvelope,
		},
	)
	if err != nil {
		return fmt.Errorf("unstaked actor router: %w", err)
	}
	return nil
}

// setupHost creates a new libp2p host and assigns it to `m.host`. Libp2p host
// starts listening upon instantiation.
func (m *p2pModule) setupHost() (err error) {
	m.logger.Debug().Msg("creating new libp2p host")

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

// handlePocketEnvelope deserializes the received `PocketEnvelope` data and publishes
// a copy of its `Content` to the application event bus.
func (m *p2pModule) handlePocketEnvelope(pocketEnvelopeBz []byte) error {
	poktEnvelope := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(pocketEnvelopeBz, &poktEnvelope); err != nil {
		return fmt.Errorf("decoding network message: %w", err)
	}

	if m.isNonceAlreadyObserved(poktEnvelope.Nonce) {
		// skip passing redundant message to application layer
		return nil
	}

	if err := m.observeNonce(poktEnvelope.Nonce); err != nil {
		return fmt.Errorf("pocket envelope nonce: %w", err)
	}

	// NB: Explicitly constructing a new `PocketEnvelope` literal with content
	// rather than forwarding `poktEnvelope` to avoid blindly passing additional
	// fields as the protobuf type changes. Additionally, strips the `Nonce` field.
	event := messaging.PocketEnvelope{
		Content: poktEnvelope.Content,
	}
	m.GetBus().PublishEventToBus(&event)
	return nil
}

// observeNonce adds the nonce to the deduper if it has not been observed.
func (m *p2pModule) observeNonce(nonce utils.Nonce) error {
	// Add the nonce to the deduper
	return m.nonceDeduper.Push(nonce)
}

// isNonceAlreadyObserved returns whether the nonce has been observed within the
// deuper's capacity of recent messages.
// DISCUSS(#278): Add more tests to verify this is sufficient for deduping purposes.
func (m *p2pModule) isNonceAlreadyObserved(nonce utils.Nonce) bool {
	if !m.nonceDeduper.Contains(nonce) {
		return false
	}

	m.logger.Debug().
		Uint64("nonce", nonce).
		Msgf("message already processed, skipping")

	m.redundantNonceTelemetry(nonce)
	return true
}

func (m *p2pModule) redundantNonceTelemetry(nonce utils.Nonce) {
	blockHeight := m.currentHeightProvider.CurrentHeight()
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			telemetry.P2P_EVENT_METRICS_NAMESPACE,
			telemetry.P2P_BROADCAST_MESSAGE_REDUNDANCY_PER_BLOCK_EVENT_METRIC_NAME,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_NONCE_LABEL, nonce,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_HEIGHT_LABEL, blockHeight,
		)
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

func (m *p2pModule) getStakedPeerstore() (typesP2P.Peerstore, error) {
	return m.pstoreProvider.GetStakedPeerstoreAtHeight(
		m.currentHeightProvider.CurrentHeight(),
	)
}

// isStakedActor returns whether the current node is a staked actor at the current height.
// Return an error if a peerstore can't be provided.
func (m *p2pModule) isStakedActor() (bool, error) {
	pstore, err := m.getStakedPeerstore()
	if err != nil {
		return false, fmt.Errorf("getting staked peerstore: %w", err)
	}

	// Ensure self address is present in current height's staked actor set.
	if self := pstore.GetPeer(m.address); self != nil {
		return true, nil
	}
	return false, nil
}
